package web

import "github.com/gin-gonic/gin"


// 全局变量的传递问题，传递到gin
func ConfigMiddle(m map[string]interface{}) gin.HandlerFunc  {
	return func(c *gin.Context) {
		for k,v :=range m {
			c.Set(k,v)
			c.Next()
		}

	}

}
