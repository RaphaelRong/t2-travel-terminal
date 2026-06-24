package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func listPlansHandler(c *gin.Context) {
	plans := []gin.H{
		{
			"id":          "free",
			"name":        "Free",
			"description": "适合个人或小型团队试用",
			"price_monthly": 0,
			"features": []string{
				"最多 3 个项目",
				"基础报表",
			},
		},
		{
			"id":          "pro",
			"name":        "Pro",
			"description": "适合成长中的团队",
			"price_monthly": 29,
			"features": []string{
				"无限项目",
				"高级报表",
				"成员邀请",
			},
		},
		{
			"id":          "enterprise",
			"name":        "Enterprise",
			"description": "适合大型企业",
			"price_monthly": 99,
			"features": []string{
				"SSO",
				"专属支持",
				"审计日志",
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{"plans": plans})
}
