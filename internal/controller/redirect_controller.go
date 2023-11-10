package controller

import (
	"context"
	"encoding/json"
	"fernandoglatz/url-management/internal/core/common/utils/constants"
	"fernandoglatz/url-management/internal/core/common/utils/exceptions"
	"fernandoglatz/url-management/internal/core/common/utils/log"
	"fernandoglatz/url-management/internal/core/entity"
	"fernandoglatz/url-management/internal/core/model/request"
	"fernandoglatz/url-management/internal/core/port/service"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

type RedirectController struct {
	service service.IRedirectService
}

func NewRedirectController(service service.IRedirectService) *RedirectController {
	return &RedirectController{
		service: service,
	}
}

// @Tags	redirect
// @Summary	Get redirects
// @Produce	json
// @Success	200	{array}		entity.Redirect
// @Failure	400	{object}	response.Response
// @Failure	500	{object}	response.Response
// @Router	/redirect [get]
func (controller *RedirectController) Get(ginCtx *gin.Context) {
	ctx := GetContext(ginCtx)
	log.Info(ctx).Msg("Getting redirects")

	redirects, err := controller.service.GetAll(ctx)
	if err != nil {
		HandleError(ctx, ginCtx, err)
		return
	}

	ginCtx.JSON(http.StatusOK, redirects)
}

// @Tags	redirect
// @Summary	Get redirect
// @Param	id		path	string  true "id"
// @Produce	json
// @Success	200	{object}	entity.Redirect
// @Failure	400	{object}	response.Response
// @Failure	500	{object}	response.Response
// @Router	/redirect/{id} [get]
func (controller *RedirectController) GetId(ginCtx *gin.Context) {
	ctx := GetContext(ginCtx)
	id := ginCtx.Param("id")

	log.Info(ctx).Msg(fmt.Sprintf("Getting redirect %s", id))

	redirect, err := controller.service.Get(ctx, id)
	if err != nil {
		HandleError(ctx, ginCtx, err)
		return
	}

	ginCtx.JSON(http.StatusOK, redirect)
}

// @Tags	redirect
// @Summary	Update redirect
// @Param	id		path	string  true "id"
// @Param	request	body	request.RedirectRequest true "body"
// @Accept	json
// @Produce	json
// @Success	200	{object}	entity.Redirect
// @Failure	400	{object}	response.Response
// @Failure	500	{object}	response.Response
// @Router		/redirect/{id} [post]
func (controller *RedirectController) Post(ginCtx *gin.Context) {
	id := ginCtx.Param(constants.ID)
	controller.save(ginCtx, &id, false)
}

// @Tags	redirect
// @Summary	Create redirect
// @Param	request	body	request.RedirectRequest true "body"
// @Accept	json
// @Produce	json
// @Success	200	{object}	entity.Redirect
// @Failure	400	{object}	response.Response
// @Failure	500	{object}	response.Response
// @Router		/redirect [put]
func (controller *RedirectController) Put(ginCtx *gin.Context) {
	controller.save(ginCtx, nil, true)
}

// @Tags	redirect
// @Summary	Update redirect
// @Param	id		path	string  true "id"
// @Param	request	body	request.RedirectRequest true "body"
// @Accept	json
// @Produce	json
// @Success	200	{object}	entity.Redirect
// @Failure	400	{object}	response.Response
// @Failure	500	{object}	response.Response
// @Router		/redirect/{id} [put]
func (controller *RedirectController) PutId(ginCtx *gin.Context) {
	id := ginCtx.Param(constants.ID)
	controller.save(ginCtx, &id, true)
}

// @Tags	redirect
// @Summary	Delete redirect
// @Param	id		path	string  true "id"
// @Produce	json
// @Success	204
// @Failure	400	{object}	response.Response
// @Failure	500	{object}	response.Response
// @Router	/redirect/{id} [delete]
func (controller *RedirectController) DeleteId(ginCtx *gin.Context) {
	ctx := GetContext(ginCtx)
	id := ginCtx.Param("id")

	log.Info(ctx).Msg(fmt.Sprintf("Removing redirect %s", id))

	redirect, err := controller.service.Get(ctx, id)
	if err != nil {
		HandleError(ctx, ginCtx, err)
		return
	}

	err = controller.service.Remove(ctx, redirect)
	if err != nil {
		HandleError(ctx, ginCtx, err)
	} else {
		ginCtx.Status(http.StatusNoContent)
	}
}

func (controller *RedirectController) save(ginCtx *gin.Context, id *string, override bool) {
	ctx := GetContext(ginCtx)

	var redirectRequest request.RedirectRequest
	var redirect entity.Redirect
	var errw *exceptions.WrappedError

	err := ginCtx.ShouldBindJSON(&redirectRequest)
	if err != nil {
		HandleError(ctx, ginCtx, &exceptions.WrappedError{
			BaseError: exceptions.InvalidJSON,
			Error:     err,
		})
		return
	}

	if id != nil {
		redirect, errw = controller.service.Get(ctx, *id)
		if errw != nil && !override {
			HandleError(ctx, ginCtx, errw)
			return
		}
		redirect.ID = *id
	}

	jsonData, _ := json.Marshal(redirectRequest)
	json.Unmarshal(jsonData, &redirect)

	errw = controller.service.Save(ctx, &redirect)
	if errw != nil {
		HandleError(ctx, ginCtx, errw)
		return

	} else {
		ginCtx.JSON(http.StatusOK, redirect)
	}
}

// @Tags	redirect
// @Summary	Execute redirect
// @Param	to		query	string  true "to"
// @Produce	json
// @Success	307
// @Failure	400	{object}	response.Response
// @Failure	500	{object}	response.Response
// @Router  / [get]
func (controller *RedirectController) Execute(ginCtx *gin.Context) {
	ctx := GetContext(ginCtx)
	id := ginCtx.Query("to")

	if len(id) > constants.ZERO {
		log.Info(ctx).Msg(fmt.Sprintf("Executing redirect %s", id))

		redirect, err := controller.service.Get(ctx, id)
		if err != nil {
			HandleError(ctx, ginCtx, err)
			return
		}

		controller.redirect(ctx, ginCtx, redirect)

	} else {
		controller.NoRoute(ginCtx)
	}
}

func (controller *RedirectController) NoRoute(ginCtx *gin.Context) {
	ctx := GetContext(ginCtx)

	host := ginCtx.Request.Host
	dns, _, _ := net.SplitHostPort(host)

	if len(dns) == constants.ZERO {
		dns = host
	}

	log.Info(ctx).Msg(fmt.Sprintf("Searching redirect for [%s]", dns))

	redirect, err := controller.service.GetByDNS(ctx, dns)
	if err == nil {
		controller.redirect(ctx, ginCtx, redirect)

	} else if err.BaseError != exceptions.RecordNotFound {
		HandleError(ctx, ginCtx, err)
	}
}

func (controller *RedirectController) redirect(ctx context.Context, ginCtx *gin.Context, redirect entity.Redirect) {
	if redirect.Proxy {
		client := &http.Client{}

		urlDestination, err := url.Parse(redirect.Destination)
		if err != nil {
			HandleError(ctx, ginCtx, &exceptions.WrappedError{Error: err})
			return
		}

		uri := ginCtx.Request.RequestURI
		body := ginCtx.Request.Body
		method := ginCtx.Request.Method
		headers := ginCtx.Request.Header

		if uri == "/" {
			uri = urlDestination.RequestURI()
		}

		destination := urlDestination.Scheme + "://" + urlDestination.Hostname()

		defer body.Close()

		request, _ := http.NewRequest(method, destination+uri, body)
		for key, values := range headers {
			value := values[constants.ZERO]

			if key == "Host" || key == "Origin" {
				value = destination

			} else if key == "Referer" {
				urlReferer, err := urlDestination.Parse(value)
				if err == nil {
					value = destination + urlReferer.Path
				}
			}

			request.Header.Set(key, value)
		}

		response, err := client.Do(request)
		if err != nil {
			HandleError(ctx, ginCtx, &exceptions.WrappedError{Error: err})
			return
		}

		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			HandleError(ctx, ginCtx, &exceptions.WrappedError{Error: err})
			return
		}

		defer response.Body.Close()

		contentType := response.Header.Get("Content-Type")

		for key, values := range response.Header {
			value := values[constants.ZERO]
			ginCtx.Header(key, value)
		}

		ginCtx.Render(response.StatusCode, render.Data{
			ContentType: contentType,
			Data:        responseBody,
		})

	} else {
		ginCtx.Redirect(http.StatusTemporaryRedirect, redirect.Destination)
	}
}
