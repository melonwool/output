package output

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var _output map[int]ErrOut

const (
	// Param 参数错误
	Param = 40000
)

type (
	// ErrOut 异常返回
	ErrOut struct {
		StatusCode   int    `json:"-"`
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	}
	// Data 带数据的结果正常返回
	Data struct {
		ErrorCode    int         `json:"error_code"`
		ErrorMessage string      `json:"error_message"`
		Data         interface{} `json:"data"`
	}
	// PageScroll 滚动翻页数据
	PageScroll struct {
		LastID interface{} `json:"last_id"`
		More   bool        `json:"more"`
		List   interface{} `json:"list"`
	}
)

func init() {
	vip := viper.New()
	vip.SetConfigType("toml")
	options := []OptFunc{
		WithAddr(os.Getenv("CONSUL_ADDR")),
		WithDC(os.Getenv("CONSUL_DC")),
		WithToken(os.Getenv("CONSUL_TOKEN")),
	}
	client := NewWatcher(options...)
	client.WatchKey("base/error", ReloadViper(vip))
}

// ErrMsg 返回异常信息
func ErrMsg(code int, c *gin.Context) {
	output := _output[code]
	if output.StatusCode == 0 {
		output = ErrOut{
			StatusCode:   500,
			ErrorCode:    5000,
			ErrorMessage: "系统异常",
		}
		c.JSON(output.StatusCode, output)
	} else {
		c.JSON(output.StatusCode, output)
	}
}

// SuccessMsg 返回正常成功的消息
func SuccessMsg(message string, c *gin.Context) {
	output := _output[0]
	// 替换为需要展示的message
	output.ErrorMessage = message
	c.JSON(200, output)
}

// SuccessData 成功返回
func SuccessData(data Data, c *gin.Context) {
	c.JSON(http.StatusOK, data)
}
