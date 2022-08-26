// ⚡️ Fiber is an Express inspired web framework written in Go with ☕️
// 📝 Github Repository: https://github.com/gofiber/fiber
// 📌 API Documentation: https://docs.gofiber.io

package fiber

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

// go test -run Test_Redirect_To
func Test_Redirect_To(t *testing.T) {
	t.Parallel()
	app := New()
	c := app.NewCtx(&fasthttp.RequestCtx{})

	c.Redirect().To("http://default.com")
	require.Equal(t, 302, c.Response().StatusCode())
	require.Equal(t, "http://default.com", string(c.Response().Header.Peek(HeaderLocation)))

	c.Redirect().Status(301).To("http://example.com")
	require.Equal(t, 301, c.Response().StatusCode())
	require.Equal(t, "http://example.com", string(c.Response().Header.Peek(HeaderLocation)))
}

// go test -run Test_Redirect_Route_WithParams
func Test_Redirect_Route_WithParams(t *testing.T) {
	t.Parallel()
	app := New()
	app.Get("/user/:name", func(c Ctx) error {
		return c.JSON(c.Params("name"))
	}).Name("user")
	c := app.NewCtx(&fasthttp.RequestCtx{})

	c.Redirect().Route("user", RedirectConfig{
		Params: Map{
			"name": "fiber",
		},
	})
	require.Equal(t, 302, c.Response().StatusCode())
	require.Equal(t, "/user/fiber", string(c.Response().Header.Peek(HeaderLocation)))
}

// go test -run Test_Redirect_Route_WithParams_WithQueries
func Test_Redirect_Route_WithParams_WithQueries(t *testing.T) {
	t.Parallel()
	app := New()
	app.Get("/user/:name", func(c Ctx) error {
		return c.JSON(c.Params("name"))
	}).Name("user")
	c := app.NewCtx(&fasthttp.RequestCtx{})

	c.Redirect().Route("user", RedirectConfig{
		Params: Map{
			"name": "fiber",
		},
		Queries: map[string]string{"data[0][name]": "john", "data[0][age]": "10", "test": "doe"},
	})
	require.Equal(t, 302, c.Response().StatusCode())
	// analysis of query parameters with url parsing, since a map pass is always randomly ordered
	location, err := url.Parse(string(c.Response().Header.Peek(HeaderLocation)))
	require.NoError(t, err, "url.Parse(location)")
	require.Equal(t, "/user/fiber", location.Path)
	require.Equal(t, url.Values{"data[0][name]": []string{"john"}, "data[0][age]": []string{"10"}, "test": []string{"doe"}}, location.Query())
}

// go test -run Test_Redirect_Route_WithOptionalParams
func Test_Redirect_Route_WithOptionalParams(t *testing.T) {
	t.Parallel()
	app := New()
	app.Get("/user/:name?", func(c Ctx) error {
		return c.JSON(c.Params("name"))
	}).Name("user")
	c := app.NewCtx(&fasthttp.RequestCtx{})

	c.Redirect().Route("user", RedirectConfig{
		Params: Map{
			"name": "fiber",
		},
	})
	require.Equal(t, 302, c.Response().StatusCode())
	require.Equal(t, "/user/fiber", string(c.Response().Header.Peek(HeaderLocation)))
}

// go test -run Test_Redirect_Route_WithOptionalParamsWithoutValue
func Test_Redirect_Route_WithOptionalParamsWithoutValue(t *testing.T) {
	t.Parallel()
	app := New()
	app.Get("/user/:name?", func(c Ctx) error {
		return c.JSON(c.Params("name"))
	}).Name("user")
	c := app.NewCtx(&fasthttp.RequestCtx{})

	c.Redirect().Route("user")
	require.Equal(t, 302, c.Response().StatusCode())
	require.Equal(t, "/user/", string(c.Response().Header.Peek(HeaderLocation)))
}

// go test -run Test_Redirect_Route_WithGreedyParameters
func Test_Redirect_Route_WithGreedyParameters(t *testing.T) {
	t.Parallel()
	app := New()
	app.Get("/user/+", func(c Ctx) error {
		return c.JSON(c.Params("+"))
	}).Name("user")
	c := app.NewCtx(&fasthttp.RequestCtx{})

	c.Redirect().Route("user", RedirectConfig{
		Params: Map{
			"+": "test/routes",
		},
	})
	require.Equal(t, 302, c.Response().StatusCode())
	require.Equal(t, "/user/test/routes", string(c.Response().Header.Peek(HeaderLocation)))
}

// go test -run Test_Redirect_Back
func Test_Redirect_Back(t *testing.T) {
	t.Parallel()
	app := New()
	app.Get("/", func(c Ctx) error {
		return c.JSON("Home")
	}).Name("home")
	c := app.NewCtx(&fasthttp.RequestCtx{})

	c.Redirect().Back("/")
	require.Equal(t, 302, c.Response().StatusCode())
	require.Equal(t, "/", string(c.Response().Header.Peek(HeaderLocation)))
}

// go test -run Test_Redirect_Back_WithReferer
func Test_Redirect_Back_WithReferer(t *testing.T) {
	t.Parallel()
	app := New()
	app.Get("/", func(c Ctx) error {
		return c.JSON("Home")
	}).Name("home")
	app.Get("/back", func(c Ctx) error {
		return c.JSON("Back")
	}).Name("back")
	c := app.NewCtx(&fasthttp.RequestCtx{})

	c.Request().Header.Set(HeaderReferer, "/back")
	c.Redirect().Back("/")
	require.Equal(t, 302, c.Response().StatusCode())
	require.Equal(t, "/back", c.Get(HeaderReferer))
	require.Equal(t, "/back", string(c.Response().Header.Peek(HeaderLocation)))
}

// go test -run Test_Redirect_Route_WithFlashMessages
func Test_Redirect_Route_WithFlashMessages(t *testing.T) {
	t.Parallel()

	app := New()
	app.Get("/user", func(c Ctx) error {
		return c.SendString("user")
	}).Name("user")

	c := app.NewCtx(&fasthttp.RequestCtx{}).(*DefaultCtx)

	c.Redirect().With("success", "1").With("message", "test").Route("user")

	require.Equal(t, 302, c.Response().StatusCode())
	require.Equal(t, "/user", string(c.Response().Header.Peek(HeaderLocation)))

	equal := string(c.Response().Header.Peek(HeaderSetCookie)) == "fiber_flash=success:1,k:message:test; path=/; SameSite=Lax" || string(c.Response().Header.Peek(HeaderSetCookie)) == "fiber_flash=message:test,k:success:1; path=/; SameSite=Lax"
	require.True(t, equal)

	c.Redirect().setFlash()
	require.Equal(t, "fiber_flash=; expires=Tue, 10 Nov 2009 23:00:00 GMT", string(c.Response().Header.Peek(HeaderSetCookie)))
}

// go test -run Test_Redirect_Route_WithOldInput
func Test_Redirect_Route_WithOldInput(t *testing.T) {
	t.Parallel()

	app := New()
	app.Get("/user", func(c Ctx) error {
		return c.SendString("user")
	}).Name("user")

	c := app.NewCtx(&fasthttp.RequestCtx{}).(*DefaultCtx)

	c.Request().URI().SetQueryString("id=1&name=tom")
	c.Redirect().With("success", "1").With("message", "test").WithInput().Route("user")

	require.Equal(t, 302, c.Response().StatusCode())
	require.Equal(t, "/user", string(c.Response().Header.Peek(HeaderLocation)))

	require.True(t, strings.Contains(string(c.Response().Header.Peek(HeaderSetCookie)), "fiber_flash="))
	require.True(t, strings.Contains(string(c.Response().Header.Peek(HeaderSetCookie)), "success:1"))
	require.True(t, strings.Contains(string(c.Response().Header.Peek(HeaderSetCookie)), "message:test"))

	require.True(t, strings.Contains(string(c.Response().Header.Peek(HeaderSetCookie)), "old_input_data_id:1"))
	require.True(t, strings.Contains(string(c.Response().Header.Peek(HeaderSetCookie)), "old_input_data_name:tom"))

	c.Redirect().setFlash()
	require.Equal(t, "fiber_flash=; expires=Tue, 10 Nov 2009 23:00:00 GMT", string(c.Response().Header.Peek(HeaderSetCookie)))
}

// go test -run Test_Redirect_setFlash
func Test_Redirect_setFlash(t *testing.T) {
	t.Parallel()

	app := New()
	app.Get("/user", func(c Ctx) error {
		return c.SendString("user")
	}).Name("user")

	c := app.NewCtx(&fasthttp.RequestCtx{}).(*DefaultCtx)

	c.Request().Header.Set(HeaderCookie, "fiber_flash=success:1,k:message:test,k:old_input_data_name:tom,k:old_input_data_id:1")

	c.Redirect().setFlash()

	require.Equal(t, "fiber_flash=; expires=Tue, 10 Nov 2009 23:00:00 GMT", string(c.Response().Header.Peek(HeaderSetCookie)))

	require.Equal(t, "1", c.Redirect().Message("success"))
	require.Equal(t, "test", c.Redirect().Message("message"))
	require.Equal(t, map[string]string{"success": "1", "message": "test"}, c.Redirect().Messages())

	require.Equal(t, "1", c.Redirect().OldInput("id"))
	require.Equal(t, "tom", c.Redirect().OldInput("name"))
	require.Equal(t, map[string]string{"id": "1", "name": "tom"}, c.Redirect().OldInputs())
}

// go test -v -run=^$ -bench=Benchmark_Redirect_Route -benchmem -count=4
func Benchmark_Redirect_Route(b *testing.B) {
	app := New()
	app.Get("/user/:name", func(c Ctx) error {
		return c.JSON(c.Params("name"))
	}).Name("user")

	c := app.NewCtx(&fasthttp.RequestCtx{}).(*DefaultCtx)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		c.Redirect().Route("user", RedirectConfig{
			Params: Map{
				"name": "fiber",
			},
		})
	}

	require.Equal(b, 302, c.Response().StatusCode())
	require.Equal(b, "/user/fiber", string(c.Response().Header.Peek(HeaderLocation)))
}

// go test -v -run=^$ -bench=Benchmark_Redirect_Route_WithQueries -benchmem -count=4
func Benchmark_Redirect_Route_WithQueries(b *testing.B) {
	app := New()
	app.Get("/user/:name", func(c Ctx) error {
		return c.JSON(c.Params("name"))
	}).Name("user")

	c := app.NewCtx(&fasthttp.RequestCtx{}).(*DefaultCtx)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		c.Redirect().Route("user", RedirectConfig{
			Params: Map{
				"name": "fiber",
			},
			Queries: map[string]string{"a": "a", "b": "b"},
		})
	}

	require.Equal(b, 302, c.Response().StatusCode())
	// analysis of query parameters with url parsing, since a map pass is always randomly ordered
	location, err := url.Parse(string(c.Response().Header.Peek(HeaderLocation)))
	require.NoError(b, err, "url.Parse(location)")
	require.Equal(b, "/user/fiber", location.Path)
	require.Equal(b, url.Values{"a": []string{"a"}, "b": []string{"b"}}, location.Query())
}

// go test -v -run=^$ -bench=Benchmark_Redirect_Route_WithFlashMessages -benchmem -count=4
func Benchmark_Redirect_Route_WithFlashMessages(b *testing.B) {
	app := New()
	app.Get("/user", func(c Ctx) error {
		return c.SendString("user")
	}).Name("user")

	c := app.NewCtx(&fasthttp.RequestCtx{}).(*DefaultCtx)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		c.Redirect().With("success", "1").With("message", "test").Route("user")
	}

	require.Equal(b, 302, c.Response().StatusCode())
	require.Equal(b, "/user", string(c.Response().Header.Peek(HeaderLocation)))

	equal := string(c.Response().Header.Peek(HeaderSetCookie)) == "fiber_flash=success:1,k:message:test; path=/; SameSite=Lax" || string(c.Response().Header.Peek(HeaderSetCookie)) == "fiber_flash=message:test,k:success:1; path=/; SameSite=Lax"
	require.True(b, equal)

	c.Redirect().setFlash()
	require.Equal(b, "fiber_flash=; expires=Tue, 10 Nov 2009 23:00:00 GMT", string(c.Response().Header.Peek(HeaderSetCookie)))
}

// go test -v -run=^$ -bench=Benchmark_Redirect_Route_WithFlashMessages -benchmem -count=4
func Benchmark_Redirect_setFlash(b *testing.B) {
	app := New()
	app.Get("/user", func(c Ctx) error {
		return c.SendString("user")
	}).Name("user")

	c := app.NewCtx(&fasthttp.RequestCtx{}).(*DefaultCtx)

	c.Request().Header.Set(HeaderCookie, "fiber_flash=success:1,k:message:test,k:old_input_data_name:tom,k:old_input_data_id:1")

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		c.Redirect().setFlash()
	}

	require.Equal(b, "fiber_flash=; expires=Tue, 10 Nov 2009 23:00:00 GMT", string(c.Response().Header.Peek(HeaderSetCookie)))

	require.Equal(b, "1", c.Redirect().Message("success"))
	require.Equal(b, "test", c.Redirect().Message("message"))
	require.Equal(b, map[string]string{"success": "1", "message": "test"}, c.Redirect().Messages())

	require.Equal(b, "1", c.Redirect().OldInput("id"))
	require.Equal(b, "tom", c.Redirect().OldInput("name"))
	require.Equal(b, map[string]string{"id": "1", "name": "tom"}, c.Redirect().OldInputs())
}
