package service

import "time"

// Product 产物管理表结构体
type Product struct {
	Id          string    `json:"id" orm:"pk;size(36);column(id);type(string);unique;description(主键ID)"`
	Name        string    `json:"name" orm:"column(name);size(255);null(false);description(产物名称)"`
	ComponentId string    `json:"component_id" orm:"column(component_id);null(false);description(关联组件ID)"`
	Size        int       `json:"size" orm:"column(size);null(false);description(产物大小)"`
	DownloadUrl string    `json:"download_url" orm:"column(download_url);size(255);null(false);description(下载地址)"`
	ArchiveTime time.Time `json:"archive_time" orm:"column(archive_time);type(datetime);null(false);description(归档时间)"`
	CreatedAt   time.Time `json:"created_at" orm:"column(created_at);type(datetime);null;description(创建时间)"`
	UpdatedAt   time.Time `json:"updated_at" orm:"column(updated_at);type(datetime);null;description(更新时间)"`
}
