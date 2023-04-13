package dao

import (
	"errors"

	"gorm.io/gorm"
)

// Place 模型定义
type Place struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Latitude    float64
	Longitude   float64
	Description string
}

// PlaceDAO 接口定义
type PlaceDAO interface {
	Create(place *Place) error
	Update(place *Place) error
	Delete(id uint) error
	GetByID(id uint) (*Place, error)
	GetAll() ([]*Place, error)
}

// PlaceDAOImpl 实现 PlaceDAO 接口
type PlaceDAOImpl struct {
	db *gorm.DB
}

// NewPlaceDAOImpl 创建一个新的 PlaceDAOImpl 实例
func NewPlaceDAOImpl(db *gorm.DB) PlaceDAO {
	return &PlaceDAOImpl{db}
}

// Create 创建一个新的地点
func (dao *PlaceDAOImpl) Create(place *Place) error {
	result := dao.db.Create(place)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Update 更新现有的地点
func (dao *PlaceDAOImpl) Update(place *Place) error {
	result := dao.db.Save(place)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}

// Delete 删除指定 ID 的地点
func (dao *PlaceDAOImpl) Delete(id uint) error {
	result := dao.db.Delete(&Place{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}

// GetByID 根据 ID 获取地点
func (dao *PlaceDAOImpl) GetByID(id uint) (*Place, error) {
	var place Place
	result := dao.db.First(&place, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &place, nil
}

// GetAll 获取所有地点
func (dao *PlaceDAOImpl) GetAll() ([]*Place, error) {
	var places []*Place
	result := dao.db.Find(&places)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return places, nil
}
