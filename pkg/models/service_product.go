package models

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/google/uuid"
	"github.com/rouroumaibing/go-devops/pkg/apis/service"
)

func init() {
	orm.RegisterModel(new(service.Product))
}

func GetProductById(id string) (*service.Product, error) {
	lock.RLock()
	defer lock.RUnlock()

	dao := orm.NewOrm()
	sql := `SELECT id, name, component_id, size, download_url, archive_time, created_at, updated_at FROM product WHERE id = ?`
	var product service.Product
	err := dao.Raw(sql, id).QueryRow(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func CreateProduct(product *service.Product) error {
	if product.Name == "" || product.ComponentId == "" || product.Size <= 0 || product.DownloadUrl == "" {
		return fmt.Errorf("产物名称、关联组件ID、产物大小和下载地址不能为空")
	}

	lock.Lock()
	defer lock.Unlock()

	productID := uuid.New().String()
	dao := orm.NewOrm()
	sql := `INSERT INTO product (id, name, component_id, size, download_url, archive_time, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err := dao.Raw(sql, productID, product.Name, product.ComponentId, product.Size, product.DownloadUrl, product.ArchiveTime).Exec()
	if err != nil {
		return fmt.Errorf("创建产物失败: %v", err)
	}
	return nil
}

func UpdateProductById(product *service.Product) error {
	sql := "UPDATE product SET updated_at = CURRENT_TIMESTAMP"
	params := []interface{}{}
	fields := []string{}

	if product.Name != "" {
		fields = append(fields, "name = ?")
		params = append(params, product.Name)
	}
	if product.ComponentId != "" {
		fields = append(fields, "component_id = ?")
		params = append(params, product.ComponentId)
	}
	if product.Size > 0 {
		fields = append(fields, "size = ?")
		params = append(params, product.Size)
	}
	if product.DownloadUrl != "" {
		fields = append(fields, "download_url = ?")
		params = append(params, product.DownloadUrl)
	}
	if !product.ArchiveTime.IsZero() {
		fields = append(fields, "archive_time = ?")
		params = append(params, product.ArchiveTime)
	}
	// 没有需要更新的字段
	if len(fields) == 0 {
		return nil
	}

	sql += ", " + strings.Join(fields, ", ") + " WHERE id = ?"
	params = append(params, product.Id)

	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	res, err := dao.Raw(sql, params...).Exec()
	if err != nil {
		return fmt.Errorf("更新产物失败: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fmt.Errorf("未找到ID为%s的产物或没有字段被更新", product.Id)
	}

	return nil
}

func DeleteProductById(id string) error {
	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	sql := `DELETE FROM product WHERE id = ?`
	_, err := dao.Raw(sql, id).Exec()
	if err != nil {
		return fmt.Errorf("删除产品失败: %v", err)
	}
	return nil
}
