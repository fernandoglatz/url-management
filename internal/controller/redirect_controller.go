package controller

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fernandoglatz/url-management/internal/core/common/utils/constants"
	"fernandoglatz/url-management/internal/core/common/utils/exceptions"
	"fernandoglatz/url-management/internal/core/common/utils/log"
	"fernandoglatz/url-management/internal/core/entity"
	redirecttype "fernandoglatz/url-management/internal/core/entity/redirect"
	"fernandoglatz/url-management/internal/core/model/request"
	"fernandoglatz/url-management/internal/core/port/service"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"fernandoglatz/url-management/internal/core/common/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

var externalURLPattern = regexp.MustCompile(`url\((['"]?)((?:https?:)?//[^'")\s,]+)(['"]?)\)`)
var externalLinkPattern = regexp.MustCompile(`(<link\b[^>]*\bhref=["'])((?:https?:)?//[^"']+)(["'])`)
var externalScriptPattern = regexp.MustCompile(`(<script\b[^>]*\bsrc=["'])((?:https?:)?//[^"']+)(["'])`)
var quotedURLPattern = regexp.MustCompile(`(?:(\w+(?::\w+)*)=)?(["'])((?:https?:)?//[^"'\\\s<>]+)(["'])`)
var externalSrcsetPattern = regexp.MustCompile(`(?i)(\bsrcset=["'])([^"']+)(["'])`)
var srcsetEntryURLPattern = regexp.MustCompile(`((?:https?:)?//[^\s,]+)`)
var integrityAttrPattern = regexp.MustCompile(`(?i)\s+integrity=["'][^"']*["']`)
var mfeRemoteURLPattern = regexp.MustCompile(`(@)((?:https?:)?//[^\s"'\\<>]+)`)
var htmlEncodedURLPattern = regexp.MustCompile(`(&quot;)((?:https?:)?//[^&\s<>]+)(&quot;)`)

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
	switch redirect.Type {
	case redirecttype.PROXY:
		urlDestination, err := url.Parse(redirect.Destination)
		if err != nil {
			HandleError(ctx, ginCtx, &exceptions.WrappedError{Error: err})
			return
		}

		if strings.EqualFold(ginCtx.Request.Header.Get("Upgrade"), "websocket") {
			controller.proxyWebSocket(ginCtx, urlDestination)
			return
		}

		client := &http.Client{}

		uri := ginCtx.Request.RequestURI
		body := ginCtx.Request.Body
		method := ginCtx.Request.Method
		headers := ginCtx.Request.Header
		domain := ginCtx.Request.Host

		scheme := "http"
		if ginCtx.Request.TLS != nil {
			scheme = "https"
		}
		if proto := ginCtx.GetHeader("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		}
		proxyBase := scheme + "://" + domain
		proxyHost, _, _ := net.SplitHostPort(domain)
		if proxyHost == "" {
			proxyHost = domain
		}
		isHTTPS := scheme == "https"

		if uri == "/" {
			uri = urlDestination.RequestURI()
		}

		destination := urlDestination.Scheme + "://" + urlDestination.Hostname()
		destinationDomain := urlDestination.Hostname()
		destinationRootDomain := utils.ExtractRootDomain(destinationDomain)

		defer body.Close()

		hopByHopHeaders := map[string]bool{
			"Connection":          true,
			"Keep-Alive":          true,
			"Proxy-Connection":    true,
			"Transfer-Encoding":   true,
			"Upgrade":             true,
			"Proxy-Authenticate":  true,
			"Proxy-Authorization": true,
			"Te":                  true,
			"Trailers":            true,
			"Accept-Encoding":     true,
		}

		request, _ := http.NewRequest(method, destination+uri, body)
		for key, values := range headers {
			if hopByHopHeaders[key] {
				continue
			}
			newValues := make([]string, 0, len(values))
			for _, value := range values {
				newValue := strings.ReplaceAll(value, domain, destinationDomain)
				newValues = append(newValues, newValue)
			}

			request.Header[key] = newValues
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

		stripResponseHeaders := map[string]bool{
			"X-Frame-Options":                     true,
			"Content-Security-Policy":             true,
			"Content-Security-Policy-Report-Only": true,
			"Strict-Transport-Security":           true,
			"Content-Encoding":                    true,
		}

		for key, values := range response.Header {
			if stripResponseHeaders[key] {
				continue
			}
			for _, value := range values {
				var newValue string
				if key == "Set-Cookie" {
					newValue = rewriteSetCookieHeader(value, destinationDomain, destinationRootDomain, proxyHost, isHTTPS)
					if newValue == "" {
						continue
					}
				} else {
					newValue = strings.ReplaceAll(value, destinationDomain, domain)
					newValue = strings.ReplaceAll(newValue, destinationRootDomain, domain)
				}
				ginCtx.Writer.Header().Add(key, newValue)
			}
		}

		if isTextBasedContent(contentType) {
			responseBodyStr := string(responseBody)
			responseBodyStr = strings.ReplaceAll(responseBodyStr, destinationDomain, domain)
			responseBodyStr = strings.ReplaceAll(responseBodyStr, destinationRootDomain, domain)
			responseBodyStr = rewriteExternalURLs(responseBodyStr, proxyBase, proxyHost, destinationRootDomain)
			responseBody = []byte(responseBodyStr)
		}

		ginCtx.Render(response.StatusCode, render.Data{
			ContentType: contentType,
			Data:        responseBody,
		})

	case redirecttype.IFRAME:
		destination := redirect.Destination
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>* { margin: 0; padding: 0; border: 0; } html, body, iframe { width: 100%%; height: 100%%; display: block; }</style>
</head>
<body>
<iframe src="%s" allowfullscreen></iframe>
</body>
</html>`, destination)
		ginCtx.Header("Content-Type", "text/html; charset=utf-8")
		ginCtx.String(http.StatusOK, html)

	default:
		ginCtx.Redirect(http.StatusTemporaryRedirect, redirect.Destination)
	}
}

func (controller *RedirectController) proxyWebSocket(ginCtx *gin.Context, destination *url.URL) {
	ctx := GetContext(ginCtx)

	upstreamHost := destination.Host
	if destination.Port() == "" {
		switch strings.ToLower(destination.Scheme) {
		case "https", "wss":
			upstreamHost = destination.Hostname() + ":443"
		default:
			upstreamHost = destination.Hostname() + ":80"
		}
	}

	var upstreamConn net.Conn
	var dialErr error
	scheme := strings.ToLower(destination.Scheme)
	if scheme == "https" || scheme == "wss" {
		upstreamConn, dialErr = tls.Dial("tcp", upstreamHost, &tls.Config{ServerName: destination.Hostname()})
	} else {
		upstreamConn, dialErr = net.Dial("tcp", upstreamHost)
	}
	if dialErr != nil {
		HandleError(ctx, ginCtx, &exceptions.WrappedError{Error: dialErr})
		return
	}
	defer upstreamConn.Close()

	uri := ginCtx.Request.RequestURI
	if uri == "/" {
		uri = destination.RequestURI()
	}

	upstreamURL := *destination
	if parsed, parseErr := url.ParseRequestURI(uri); parseErr == nil {
		upstreamURL.Path = parsed.Path
		upstreamURL.RawPath = parsed.RawPath
		upstreamURL.RawQuery = parsed.RawQuery
	}

	upstreamReq, _ := http.NewRequest(ginCtx.Request.Method, upstreamURL.String(), ginCtx.Request.Body)
	upstreamReq.Host = destination.Host

	domain := ginCtx.Request.Host
	wsHopByHop := map[string]bool{
		"Keep-Alive":          true,
		"Proxy-Connection":    true,
		"Transfer-Encoding":   true,
		"Proxy-Authenticate":  true,
		"Proxy-Authorization": true,
		"Te":                  true,
		"Trailers":            true,
	}
	for key, values := range ginCtx.Request.Header {
		if wsHopByHop[key] {
			continue
		}
		newValues := make([]string, 0, len(values))
		for _, v := range values {
			newValues = append(newValues, strings.ReplaceAll(v, domain, destination.Host))
		}
		upstreamReq.Header[key] = newValues
	}

	if err := upstreamReq.Write(upstreamConn); err != nil {
		HandleError(ctx, ginCtx, &exceptions.WrappedError{Error: err})
		return
	}

	upstreamReader := bufio.NewReader(upstreamConn)
	upstreamResp, err := http.ReadResponse(upstreamReader, upstreamReq)
	if err != nil {
		HandleError(ctx, ginCtx, &exceptions.WrappedError{Error: err})
		return
	}
	if upstreamResp.StatusCode != http.StatusSwitchingProtocols {
		HandleError(ctx, ginCtx, &exceptions.WrappedError{
			Error: fmt.Errorf("WebSocket upstream returned %d, expected 101", upstreamResp.StatusCode),
		})
		return
	}

	hijacker, ok := ginCtx.Writer.(http.Hijacker)
	if !ok {
		HandleError(ctx, ginCtx, &exceptions.WrappedError{
			Error: fmt.Errorf("response writer does not support hijacking"),
		})
		return
	}
	clientConn, clientBuf, err := hijacker.Hijack()
	if err != nil {
		HandleError(ctx, ginCtx, &exceptions.WrappedError{Error: err})
		return
	}
	defer clientConn.Close()

	if err := upstreamResp.Write(clientBuf); err != nil {
		log.Error(ctx).Msg("Failed to write WebSocket 101 response to client: " + err.Error())
		return
	}
	if err := clientBuf.Flush(); err != nil {
		log.Error(ctx).Msg("Failed to flush WebSocket 101 response to client: " + err.Error())
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(upstreamConn, clientBuf)
		upstreamConn.Close()
	}()

	go func() {
		defer wg.Done()
		io.Copy(clientConn, upstreamReader)
		clientConn.Close()
	}()

	wg.Wait()
}

func (controller *RedirectController) CDN(ginCtx *gin.Context) {
	ctx := GetContext(ginCtx)
	targetURL := ginCtx.Query("url")

	if len(targetURL) == 0 {
		ginCtx.Status(http.StatusBadRequest)
		return
	}

	controller.serveCDN(ctx, ginCtx, targetURL)
}

// CDNPath proxies an external resource identified by /__cdnp/{host}{/path}.
// Unlike CDN (which uses a ?url= query parameter), the target host and path are
// encoded in the URL path itself so that webpack's auto publicPath detection
// strips to the correct directory when computing relative chunk URLs.
func (controller *RedirectController) CDNPath(ginCtx *gin.Context) {
	ctx := GetContext(ginCtx)

	fullpath := strings.TrimPrefix(ginCtx.Param("fullpath"), "/")
	slashIdx := strings.IndexByte(fullpath, '/')
	if slashIdx < 0 {
		ginCtx.Status(http.StatusBadRequest)
		return
	}

	host := fullpath[:slashIdx]
	pathPart := fullpath[slashIdx:]

	rawQuery := ginCtx.Request.URL.RawQuery
	targetURL := "https://" + host + pathPart
	if rawQuery != "" {
		targetURL += "?" + rawQuery
	}

	controller.serveCDN(ctx, ginCtx, targetURL)
}

func (controller *RedirectController) serveCDN(ctx context.Context, ginCtx *gin.Context, targetURL string) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		HandleError(ctx, ginCtx, &exceptions.WrappedError{Error: err})
		return
	}

	req.Header.Set("User-Agent", ginCtx.GetHeader("User-Agent"))
	if accept := ginCtx.GetHeader("Accept"); accept != "" {
		req.Header.Set("Accept", accept)
	}
	if lang := ginCtx.GetHeader("Accept-Language"); lang != "" {
		req.Header.Set("Accept-Language", lang)
	}
	if cookie := ginCtx.GetHeader("Cookie"); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	if parsed, parseErr := url.Parse(targetURL); parseErr == nil && parsed.Host != "" {
		req.Header.Set("Referer", parsed.Scheme+"://"+parsed.Host+"/")
	}

	response, err := client.Do(req)
	if err != nil {
		log.Error(ctx).Msg("CDN proxy fetch error for " + targetURL + ": " + err.Error())
		ginCtx.Status(http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error(ctx).Msg("CDN proxy read error for " + targetURL + ": " + err.Error())
		ginCtx.Status(http.StatusBadGateway)
		return
	}

	contentType := response.Header.Get("Content-Type")

	if isCSSContent(contentType) {
		scheme := "http"
		if ginCtx.Request.TLS != nil {
			scheme = "https"
		}
		if proto := ginCtx.GetHeader("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		}
		host := ginCtx.Request.Host
		proxyBase := scheme + "://" + host
		proxyHost, _, _ := net.SplitHostPort(host)
		if proxyHost == "" {
			proxyHost = host
		}
		body = []byte(rewriteExternalURLs(string(body), proxyBase, proxyHost, ""))
	}

	ginCtx.Header("Access-Control-Allow-Origin", "*")
	if cc := response.Header.Get("Cache-Control"); cc != "" {
		ginCtx.Header("Cache-Control", cc)
	}

	ginCtx.Render(response.StatusCode, render.Data{
		ContentType: contentType,
		Data:        body,
	})
}

func rewriteExternalURLs(content, proxyBase, proxyHost, destinationRootDomain string) string {
	cdnURL := func(externalURL string) string {
		normalizedURL := externalURL
		if strings.HasPrefix(externalURL, "//") {
			normalizedURL = "https:" + externalURL
		}

		parsed, err := url.Parse(normalizedURL)
		if err != nil || parsed.Host == "" {
			return externalURL
		}
		hostname := parsed.Hostname()

		// Reject non-hostname paths like //api/endpoint (no dot = not an external host)
		if !strings.Contains(hostname, ".") {
			return externalURL
		}

		// Exact proxy host — normalize to the proxy's scheme/host so that https://
		// references don't cause SSL errors when the proxy is running over HTTP.
		if hostname == proxyHost {
			return proxyBase + parsed.RequestURI()
		}

		// Subdomain of proxy host — the root domain replacement incorrectly rewrote a CDN
		// subdomain (e.g. web-assets.strava.com → web-assets.strava.fernandoglatz.com:8080).
		// Reverse it back to the original hostname and route through /__cdnp.
		if destinationRootDomain != "" && strings.HasSuffix(hostname, "."+proxyHost) {
			subdomain := hostname[:len(hostname)-len("."+proxyHost)]
			reversed, _ := url.Parse(normalizedURL)
			if reversed != nil {
				reversed.Host = subdomain + "." + destinationRootDomain
				return proxyBase + "/__cdnp/" + reversed.Hostname() + reversed.RequestURI()
			}
		}

		return proxyBase + "/__cdnp/" + parsed.Hostname() + parsed.RequestURI()
	}

	content = externalURLPattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := externalURLPattern.FindStringSubmatch(match)
		if len(sub) < 4 {
			return match
		}
		quote, externalURL := sub[1], sub[2]
		return "url(" + quote + cdnURL(externalURL) + quote + ")"
	})

	content = externalLinkPattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := externalLinkPattern.FindStringSubmatch(match)
		if len(sub) < 4 {
			return match
		}
		prefix, externalURL, quote := sub[1], sub[2], sub[3]
		return prefix + cdnURL(externalURL) + quote
	})

	content = externalScriptPattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := externalScriptPattern.FindStringSubmatch(match)
		if len(sub) < 4 {
			return match
		}
		prefix, externalURL, quote := sub[1], sub[2], sub[3]
		return prefix + cdnURL(externalURL) + quote
	})

	// Catch-all for quoted URLs not covered by the specific patterns above
	// (e.g. <img src>, JS strings, JSON values, inline styles).
	content = quotedURLPattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := quotedURLPattern.FindStringSubmatch(match)
		if len(sub) < 5 || sub[2] != sub[4] {
			return match
		}
		attrName, openQuote, externalURL := sub[1], sub[2], sub[3]
		// xmlns attributes are XML namespace identifiers, not fetchable URLs
		if strings.HasPrefix(attrName, "xmlns") {
			return match
		}
		newURL := cdnURL(externalURL)
		if newURL == externalURL {
			return match
		}
		prefix := ""
		if attrName != "" {
			prefix = attrName + "="
		}
		return prefix + openQuote + newURL + openQuote
	})

	// srcset attributes (img/source) contain space-separated "URL descriptor" entries
	// separated by commas; quotedURLPattern can't match them because URLs are followed
	// by a space+descriptor before the closing quote. Handle both srcset and srcSet (React).
	content = externalSrcsetPattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := externalSrcsetPattern.FindStringSubmatch(match)
		if len(sub) < 4 {
			return match
		}
		prefix, srcsetValue, closingQuote := sub[1], sub[2], sub[3]
		newValue := srcsetEntryURLPattern.ReplaceAllStringFunc(srcsetValue, func(u string) string {
			return cdnURL(u)
		})
		return prefix + newValue + closingQuote
	})

	// URLs inside HTML attribute values with entity-encoded quotes (&quot;...&quot;)
	// e.g. data-react-props='{"url":"https://..."}'. The quotedURLPattern only matches
	// real " or ' characters, so these are invisible to it.
	// Only rewrite proxy-subdomain URLs (produced by the earlier text replacement) to
	// avoid corrupting third-party URLs that use JSON & escapes for & in query params.
	content = htmlEncodedURLPattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := htmlEncodedURLPattern.FindStringSubmatch(match)
		if len(sub) < 4 {
			return match
		}
		parsed, parseErr := url.Parse(sub[2])
		if parseErr != nil || parsed.Host == "" {
			return match
		}
		hostname := parsed.Hostname()
		if destinationRootDomain == "" || !strings.HasSuffix(hostname, "."+proxyHost) {
			return match
		}
		newURL := cdnURL(sub[2])
		if newURL == sub[2] {
			return match
		}
		return sub[1] + newURL + sub[3]
	})

	// Module Federation remote entry format: "scope@https://..." — the URL follows "@"
	// and is not at the start of the quoted string so quotedURLPattern misses it.
	content = mfeRemoteURLPattern.ReplaceAllStringFunc(content, func(match string) string {
		sub := mfeRemoteURLPattern.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		newURL := cdnURL(sub[2])
		if newURL == sub[2] {
			return match
		}
		return sub[1] + newURL
	})

	// Strip SRI integrity attributes: proxied content is served through /__cdn which may
	// modify text content (URL rewriting), so the original hash will never match.
	content = integrityAttrPattern.ReplaceAllString(content, "")

	return content
}

func rewriteSetCookieHeader(value, destinationDomain, destinationRootDomain, proxyHost string, isHTTPS bool) string {
	parts := strings.Split(value, ";")
	if len(parts) == 0 {
		return value
	}

	nameVal := strings.TrimSpace(parts[0])
	cookieName := nameVal
	if idx := strings.IndexByte(nameVal, '='); idx >= 0 {
		cookieName = nameVal[:idx]
	}
	if !isHTTPS && (strings.HasPrefix(cookieName, "__Host-") || strings.HasPrefix(cookieName, "__Secure-")) {
		return ""
	}

	hasSecure := false
	for _, part := range parts[1:] {
		if strings.EqualFold(strings.TrimSpace(part), "secure") {
			hasSecure = true
			break
		}
	}

	result := make([]string, 0, len(parts))
	result = append(result, parts[0])

	for _, part := range parts[1:] {
		trimmed := strings.TrimSpace(part)
		lower := strings.ToLower(trimmed)

		switch {
		case strings.HasPrefix(lower, "domain="):
			domainVal := strings.TrimPrefix(trimmed[len("domain="):], ".")
			if strings.Contains(domainVal, destinationDomain) {
				domainVal = strings.ReplaceAll(domainVal, destinationDomain, proxyHost)
			} else if destinationRootDomain != "" && strings.Contains(domainVal, destinationRootDomain) {
				domainVal = strings.ReplaceAll(domainVal, destinationRootDomain, proxyHost)
			}
			result = append(result, " Domain="+domainVal)

		case lower == "secure":
			if isHTTPS {
				result = append(result, " Secure")
			}

		case strings.HasPrefix(lower, "samesite="):
			if !isHTTPS && strings.TrimPrefix(lower, "samesite=") == "none" && hasSecure {
				result = append(result, " SameSite=Lax")
			} else {
				result = append(result, " "+trimmed)
			}

		default:
			result = append(result, " "+trimmed)
		}
	}

	return strings.Join(result, ";")
}

func isCSSContent(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "text/css")
}

func isTextBasedContent(contentType string) bool {
	textTypes := []string{
		"text/html",
		"text/css",
		"text/javascript",
		"application/javascript",
		"application/x-javascript",
		"text/plain",
		"application/json",
		"application/xml",
		"text/xml",
		"application/xhtml+xml",
		"text/csv",
		"application/rss+xml",
		"application/atom+xml",
	}

	contentTypeLower := strings.ToLower(contentType)
	for _, textType := range textTypes {
		if strings.Contains(contentTypeLower, textType) {
			return true
		}
	}

	return false
}
