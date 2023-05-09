package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func main() {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	dsn := "root:123456@tcp(127.0.0.1:3306)/crud-list?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 表名变回单数
		},
	})

	fmt.Println(db)
	fmt.Println(err)

	// 连接池
	sqlDB, err := db.DB()

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(10 * time.Second) // 10秒

	// 结构体（写入数据）
	type List struct {
		gorm.Model
		Name    string `gorm: "type:varchar(20); not null" json:"name" binding:"required"`
		State   string `gorm: "type:varchar(20); not null" json:"state" binding:"required"`
		Phone   string `gorm: "type:varchar(20); not null" json:"phone" binding:"required"`
		Email   string `gorm: "type:varchar(40); not null" json:"email" binding:"required"`
		Address string `gorm: "type:varchar(200); not null" json:"address" binding:"required"`
	}

	// 迁移
	db.AutoMigrate(&List{})

	// 接口
	r := gin.Default()

	/* 测试
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "请求成功",
		})
	})
	*/

	// 增
	r.POST("/user/add", func(c *gin.Context) {
		var data List

		err := c.ShouldBindJSON(&data)

		// 判断绑定是否有错
		if err != nil {
			c.JSON(200, gin.H{
				"msg":  "添加失败",
				"data": gin.H{},
				"code": 400,
			})
		} else {

			// 数据库操作
			db.Create(&data) // 创建一条数据

			c.JSON(200, gin.H{
				"msg":  "添加成功",
				"data": data,
				"code": 200,
			})
		}
	})

	// 删
	// 1. 找到对应的id对应的条目
	// 2. 查找id
	// 3. 从数据库中删除
	// 4. 返回，id没有找到
	// restful 编码规范风格

	r.DELETE("/user/delete/:id", func(c *gin.Context) {
		var data []List

		// 接受 id
		id := c.Param("id")

		// 查找 id
		db.Where("id = ?", id).Find(&data)

		// 存在删，不存在报错
		if len(data) == 0 {
			c.JSON(200, gin.H{
				"msg":  "id没有找到，删除失败",
				"code": 400,
			})
		} else {
			// 操作数据库删除
			db.Where("id = ?", id).Delete(&data)

			c.JSON(200, gin.H{
				"msg":  "删除成功",
				"code": 200,
			})
		}
	})

	// 改
	// 1. 找到对应的id对应的条目
	// 2. 查找id
	// 3. 修改对应条目
	// 4. 返回id，找不到

	r.PUT("/user/update/:id", func(c *gin.Context) {
		var data List

		// 接受 id
		id := c.Param("id")

		// 查找 id
		db.Where("id = ?", id).Find(&data)

		// 判断是否有id
		if data.ID == 0 {
			c.JSON(200, gin.H{
				"msg":  "用户没有找到",
				"code": 400,
			})
		} else {
			err := c.ShouldBindJSON(&data)

			if err != nil {
				c.JSON(200, gin.H{
					"msg":  "修改失败",
					"code": 400,
				})
			} else {

				// 操作数据库删除
				db.Where("id = ?", id).Updates(&data)

				c.JSON(200, gin.H{
					"msg":  "修改成功",
					"code": 200,
				})
			}
		}
	})

	// 查（条件查询，全部查询/分页查询）

	// 条件查询
	r.GET("/user/list/:name", func(c *gin.Context) {

		// 获得路径参数
		name := c.Param("name")

		var dataList []List

		// 查询数据库
		db.Where("name = ?", name).Find((&dataList))

		// 判断是否查询到数据
		if len(dataList) == 0 {
			c.JSON(200, gin.H{
				"msg":  "没有查询到数据",
				"code": 400,
				"data": gin.H{},
			})
		} else {
			c.JSON(200, gin.H{
				"msg":  "查询成功",
				"code": 200,
				"data": dataList,
			})
		}
	})

	// 全部查询
	r.GET("/user/list", func(c *gin.Context) {
		var dataList []List

		// 1. 查询全部数据，分页数据
		pageNum, _ := strconv.Atoi(c.Query("pageNum"))
		pageSize, _ := strconv.Atoi(c.Query("pageSize"))

		// 判断是否需要分页
		if pageSize == 0 {
			pageSize = -1
		}
		if pageNum == 0 {
			pageNum = -1
		}

		offsetVal := (pageNum - 1) * pageSize // 分页固定写法
		if pageNum == -1 && pageSize == -1 {
			offsetVal = -1
		}

		// 返回一个总数
		var total int64
		// 查询数据库
		db.Model(dataList).Count(&total).Limit(pageSize).Offset(offsetVal).Find(&dataList)

		if len(dataList) == 0 {
			c.JSON(200, gin.H{
				"msg":  "没有查到数据",
				"code": 400,
				"data": gin.H{},
			})
		} else {
			c.JSON(200, gin.H{
				"msg":  "查询成功",
				"code": 200,
				"data": gin.H{
					"list":     dataList,
					"total":    total,
					"pageNum":  "pageNum",
					"pageSize": "pageSize",
				},
			})
		}
	})

	// 端口号
	PORT := "3001"
	r.Run(":" + PORT)

}
