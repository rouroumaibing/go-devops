package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/rouroumaibing/go-devops/pkg/apis/service"
	"github.com/rouroumaibing/go-devops/pkg/models"
	"github.com/rouroumaibing/go-devops/pkg/utils/errcode"
	"github.com/rouroumaibing/go-devops/pkg/utils/response"
)

type ProductController struct {
	beego.Controller
}

func (p *ProductController) GetProductById() {
	productId := p.Ctx.Input.Param(":id")
	if productId == "" {
		response.SetErrorResponse(p.Ctx, errcode.E.Product.ParamError.WithMessage("get product failed: "+fmt.Errorf("product ID error").Error()))
		return
	}

	product, err := models.GetProductById(productId)
	if err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Product.NotFound.WithMessage("get product failed: "+err.Error()))
		return
	}
	productJson, err := json.Marshal(product)
	if err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Product.QueryFailed.WithMessage("get product failed: "+err.Error()))
		return
	}

	response.HttpResponse(http.StatusOK, productJson, p.Ctx)

}

func (p *ProductController) CreateProduct() {
	var product service.Product
	if err := json.Unmarshal(p.Ctx.Input.RequestBody, &product); err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Product.ParamError.WithMessage("parameter error: "+err.Error()))
		return
	}
	if err := models.CreateProduct(&product); err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Product.CreateFailed.WithMessage("create product failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, p.Ctx)
}

func (p *ProductController) UpdateProductById() {
	productID := p.Ctx.Input.Param(":id")
	if productID == "" {
		response.SetErrorResponse(p.Ctx, errcode.E.Product.UpdateFailed.WithMessage("update product failed: "+fmt.Errorf("product ID error").Error()))
		return
	}

	_, err := models.GetProductById(productID)
	if err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Product.UpdateFailed.WithMessage("update product failed: "+err.Error()))
		return
	}
	var updateProduct service.Product
	if err := json.Unmarshal(p.Ctx.Input.RequestBody, &updateProduct); err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Product.UpdateFailed.WithMessage("update product failed: "+err.Error()))
		return
	}
	if err := models.UpdateProductById(&updateProduct); err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Product.UpdateFailed.WithMessage("update product failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, p.Ctx)
}

func (p *ProductController) DeleteProductById() {
	productID := p.Ctx.Input.Param(":id")
	if productID == "" {
		response.SetErrorResponse(p.Ctx, errcode.E.Product.DeleteFailed.WithMessage("delete product failed: "+fmt.Errorf("product ID error").Error()))
		return
	}
	_, err := models.GetProductById(productID)
	if err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Product.DeleteFailed.WithMessage("delete product failed: "+err.Error()))
		return
	}
	if err := models.DeleteProductById(productID); err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Product.DeleteFailed.WithMessage("delete product failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, p.Ctx)
}
