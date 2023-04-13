package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"operateExcel/dao"
	"os"
	"path/filepath"
	"strconv"

	"time"

	"github.com/tealeg/xlsx"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// PlaceHandler 包含 Place 相关的路由处理程序
type PlaceHandler struct {
	dao dao.PlaceDAO
}

// NewPlaceHandler 创建一个新的 PlaceHandler 实例
func NewPlaceHandler(dao dao.PlaceDAO) *PlaceHandler {
	return &PlaceHandler{dao}
}

// AddPlace 添加一个新的地点
func (h *PlaceHandler) AddPlace(c *gin.Context) {
	var place dao.Place
	place.Name = c.PostForm("name")
	place.Longitude, _ = strconv.ParseFloat(c.PostForm("lon"), 64)
	place.Latitude, _ = strconv.ParseFloat(c.PostForm("lat"), 64)
	place.Description = c.PostForm("desc")

	if err := h.dao.Create(&place); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/places")
}

// 导出
func (h *PlaceHandler) Export(c *gin.Context) {
	// 获取所有地点
	places, err := h.dao.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	// 创建 Excel 文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Places")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Excel file"})
		return
	}

	// 添加表头
	headerRow := sheet.AddRow()
	headerRow.AddCell().Value = "ID"
	headerRow.AddCell().Value = "Name"
	headerRow.AddCell().Value = "Lat"
	headerRow.AddCell().Value = "Lon"
	headerRow.AddCell().Value = "Description"

	// 添加数据
	for _, place := range places {
		row := sheet.AddRow()
		row.AddCell().SetValue(place.ID)
		row.AddCell().SetValue(place.Name)
		row.AddCell().SetValue(place.Latitude)
		row.AddCell().SetValue(place.Longitude)
		row.AddCell().SetValue(place.Description)
	}

	// 保存 Excel 文件到临时目录
	fileName := fmt.Sprintf("places-%d.xlsx", time.Now().Unix())
	filePath := filepath.Join(os.TempDir(), fileName)
	if err := file.Save(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save Excel file"})
		return
	}

	// 返回 Excel 文件下载链接
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.File(filePath)

}

func (h *PlaceHandler) ToImport(c *gin.Context) {
	c.HTML(http.StatusOK, "import.html", nil)
}

// 导入
// 导入
func (h *PlaceHandler) ImportPlaces(c *gin.Context) {
	// 从请求中获取上传的 Excel 文件
	file, err := c.FormFile("file")
	if err != nil {
		c.HTML(http.StatusBadRequest, "import.html", gin.H{"error": "Failed to get file"})
		return
	}

	f, _ := file.Open()

	// 读取文件内容到字节数组中
	fileBytes, err := ioutil.ReadAll(f)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "import.html", gin.H{"error": "Failed to read file"})
		return
	}

	// 将字节数组传递给xlsx.OpenBinary方法
	xlFile, err := xlsx.OpenBinary(fileBytes)
	if err != nil {
		c.HTML(http.StatusBadRequest, "import.html", gin.H{"error": "Failed to open Excel file"})
		return
	}

	// 遍历每个工作表的每一行数据，并将数据保存到数据库中
	for _, sheet := range xlFile.Sheets {
		for i, row := range sheet.Rows {
			// 跳过第一行（标题行）
			if i == 0 {
				continue
			}
			lat, _ := strconv.ParseFloat(row.Cells[2].Value, 64)
			lon, _ := strconv.ParseFloat(row.Cells[3].Value, 64)
			place := &dao.Place{
				Name:        row.Cells[1].Value,
				Latitude:    lat,
				Longitude:   lon,
				Description: row.Cells[4].Value,
			}
			if err := h.dao.Create(place); err != nil {
				c.HTML(http.StatusInternalServerError, "import.html", gin.H{"error": "Failed to save data"})
				return
			}
		}
	}

	c.Redirect(http.StatusMovedPermanently, "/places")
}

// ListPlaces 显示所有地点的列表
func (h *PlaceHandler) ListPlaces(c *gin.Context) {
	places, err := h.dao.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "index.html", gin.H{"places": places})
}

func main() {
	// 创建一个 GORM 数据库实例
	db, err := gorm.Open(mysql.Open("root:liu123456@tcp(localhost:3306)/golang_tutorial"), &gorm.Config{})
	if err != nil {
		panic("数据库连接错误")
	}
	// 自动迁移数据库，创建对应的表
	db.AutoMigrate(&dao.Place{})

	// 创建一个 Gin 实例
	r := gin.Default()

	// 创建一个 PlaceDAOImpl 实例
	dao := dao.NewPlaceDAOImpl(db)

	// 创建一个 PlaceHandler 实例
	handler := NewPlaceHandler(dao)

	// 添加一个新的地点
	r.POST("/places", handler.AddPlace)

	// 显示所有地点的列表
	r.GET("/places", handler.ListPlaces)
	r.POST("/import", handler.ImportPlaces)
	r.GET("/export", handler.Export)
	r.GET("/to_import", handler.ToImport)

	// 模板渲染
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// 启动服务器
	r.Run(":8080")
}
