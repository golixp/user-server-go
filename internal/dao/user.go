package dao

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	cacheBase "github.com/go-dev-frame/sponge/pkg/cache"
	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"user-server-go/internal/cache"
	"user-server-go/internal/database"
	"user-server-go/internal/model"
)

var _ UserDao = (*userDao)(nil)

// UserDao defining the dao interface
type UserDao interface {
	Create(ctx context.Context, table *model.User) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.User) error
	GetByID(ctx context.Context, id uint64) (*model.User, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.User, int64, error)

	DeleteByIDs(ctx context.Context, ids []uint64) error
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.User, error)
	GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.User, error)
	GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.User, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.User) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.User) error

	GetByUsername(ctx context.Context, username string) (*model.User, error)
}

type userDao struct {
	db    *gorm.DB
	cache cache.UserCache     // if nil, the cache is not used.
	sfg   *singleflight.Group // if cache is nil, the sfg is not used.
}

// NewUserDao creating the dao interface
func NewUserDao(db *gorm.DB, xCache cache.UserCache) UserDao {
	if xCache == nil {
		return &userDao{db: db}
	}
	return &userDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *userDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		data, err := d.GetByID(ctx, id)
		if err != nil {
			return err
		}
		err = d.deleteCacheByUsername(ctx, data.Username)
		if err != nil {
			return err
		}
		err = d.deleteCacheById(ctx, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *userDao) deleteCacheById(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

func (d *userDao) deleteCacheByUsername(ctx context.Context, username string) error {
	if d.cache != nil {
		return d.cache.DelByUsername(ctx, username)
	}
	return nil
}

// Create a record, insert the record and the id value is written back to the table
func (d *userDao) Create(ctx context.Context, table *model.User) error {
	_ = d.deleteCacheByUsername(ctx, table.Username)
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a record by id
func (d *userDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a record by id
func (d *userDao) UpdateByID(ctx context.Context, table *model.User) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *userDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.User) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.Username != "" {
		update["username"] = table.Username
	}
	if table.Nickname != "" {
		update["nickname"] = table.Nickname
	}
	if table.Password != "" {
		update["password"] = table.Password
	}
	if table.LoginAt.IsZero() == false {
		update["login_at"] = table.LoginAt
	}
	if table.LoginIP != "" {
		update["login_ip"] = table.LoginIP
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a record by id
func (d *userDao) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	// no cache
	if d.cache == nil {
		record := &model.User{}
		err := d.db.WithContext(ctx).Where("id = ?", id).First(record).Error
		return record, err
	}

	// get from cache
	record, err := d.cache.Get(ctx, id)
	if err == nil {
		return record, nil
	}

	// get from database
	if errors.Is(err, database.ErrCacheNotFound) {
		// for the same id, prevent high concurrent simultaneous access to database
		val, err, _ := d.sfg.Do(utils.Uint64ToStr(id), func() (interface{}, error) {
			table := &model.User{}
			err = d.db.WithContext(ctx).Where("id = ?", id).First(table).Error
			if err != nil {
				// set placeholder cache to prevent cache penetration, default expiration time 10 minutes
				if errors.Is(err, database.ErrRecordNotFound) {
					if err = d.cache.SetPlaceholder(ctx, id); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("id", id))
					}
					return nil, database.ErrRecordNotFound
				}
				return nil, err
			}
			// set cache
			if err = d.cache.Set(ctx, id, table, cache.UserExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.User)
		if !ok {
			return nil, database.ErrRecordNotFound
		}
		return table, nil
	}

	if d.cache.IsPlaceholderErr(err) {
		return nil, database.ErrRecordNotFound
	}

	return nil, err
}

// GetByColumns get paging records by column information,
// Note: query performance degrades when table rows are very large because of the use of offset.
//
// params includes paging parameters and query parameters
// paging parameters (required):
//
//	page: page number, starting from 0
//	limit: lines per page
//	sort: sort fields, default is id backwards, you can add - sign before the field to indicate reverse order, no - sign to indicate ascending order, multiple fields separated by comma
//
// query parameters (not required):
//
//	name: column name
//	exp: expressions, which default is "=",  support =, !=, >, >=, <, <=, like, in, notin, isnull, isnotnull
//	value: column value, if exp=in, multiple values are separated by commas
//	logic: logical type, default value is "and", support &, and, ||, or
//
// example: search for a male over 20 years of age
//
//	params = &query.Params{
//	    Page: 0,
//	    Limit: 20,
//	    Columns: []query.Column{
//		{
//			Name:    "age",
//			Exp: ">",
//			Value:   20,
//		},
//		{
//			Name:  "gender",
//			Value: "male",
//		},
//	}
func (d *userDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.User, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions()
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.User{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.User{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// DeleteByIDs delete records by batch id
func (d *userDao) DeleteByIDs(ctx context.Context, ids []uint64) error {
	err := d.db.WithContext(ctx).Where("id IN (?)", ids).Delete(&model.User{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, id := range ids {
		_ = d.deleteCache(ctx, id)
	}

	return nil
}

// GetByCondition get a record by condition
// query conditions:
//
//	name: column name
//	exp: expressions, which default is "=",  support =, !=, >, >=, <, <=, like, in, notin, isnull, isnotnull
//	value: column value, if exp=in, multiple values are separated by commas
//	logic: logical type, default value is "and", support &, and, ||, or
//
// example: find a male aged 20
//
//	condition = &query.Conditions{
//	    Columns: []query.Column{
//		{
//			Name:    "age",
//			Value:   20,
//		},
//		{
//			Name:  "gender",
//			Value: "male",
//		},
//	}
func (d *userDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.User, error) {
	queryStr, args, err := c.ConvertToGorm()
	if err != nil {
		return nil, err
	}

	table := &model.User{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetByIDs get records by batch id
func (d *userDao) GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.User, error) {
	// no cache
	if d.cache == nil {
		var records []*model.User
		err := d.db.WithContext(ctx).Where("id IN (?)", ids).Find(&records).Error
		if err != nil {
			return nil, err
		}
		itemMap := make(map[uint64]*model.User)
		for _, record := range records {
			itemMap[record.ID] = record
		}
		return itemMap, nil
	}

	// get form cache
	itemMap, err := d.cache.MultiGet(ctx, ids)
	if err != nil {
		return nil, err
	}

	var missedIDs []uint64
	for _, id := range ids {
		if _, ok := itemMap[id]; !ok {
			missedIDs = append(missedIDs, id)
		}
	}

	// get missed data
	if len(missedIDs) > 0 {
		// find the id of an active placeholder, i.e. an id that does not exist in database
		var realMissedIDs []uint64
		for _, id := range missedIDs {
			_, err = d.cache.Get(ctx, id)
			if d.cache.IsPlaceholderErr(err) {
				continue
			}
			realMissedIDs = append(realMissedIDs, id)
		}

		// get missed id from database
		if len(realMissedIDs) > 0 {
			var records []*model.User
			var recordIDMap = make(map[uint64]struct{})
			err = d.db.WithContext(ctx).Where("id IN (?)", realMissedIDs).Find(&records).Error
			if err != nil {
				return nil, err
			}
			if len(records) > 0 {
				for _, record := range records {
					itemMap[record.ID] = record
					recordIDMap[record.ID] = struct{}{}
				}
				if err = d.cache.MultiSet(ctx, records, cache.UserExpireTime); err != nil {
					logger.Warn("cache.MultiSet error", logger.Err(err), logger.Any("ids", records))
				}
				if len(records) == len(realMissedIDs) {
					return itemMap, nil
				}
			}
			for _, id := range realMissedIDs {
				if _, ok := recordIDMap[id]; !ok {
					if err = d.cache.SetPlaceholder(ctx, id); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("id", id))
					}
				}
			}
		}
	}

	return itemMap, nil
}

// GetByLastID get paging records by last id and limit
func (d *userDao) GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.User, error) {
	page := query.NewPage(0, limit, sort)

	records := []*model.User{}
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Limit()).Where("id < ?", lastID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *userDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.User) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *userDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	update := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	err := tx.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(update).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *userDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.User) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

// GetByUsername get a record by username
func (d *userDao) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	// no cache
	if d.cache == nil {
		record := &model.User{}
		err := d.db.WithContext(ctx).Where("username = ?", username).First(record).Error
		return record, err
	}

	// get from cache
	record, err := d.cache.GetByUsername(ctx, username)
	if err == nil {
		return record, nil
	}

	if errors.Is(err, cacheBase.CacheNotFound) {
		val, err, _ := d.sfg.Do(username, func() (interface{}, error) {
			table := &model.User{}
			err = d.db.WithContext(ctx).Where("username = ?", username).First(table).Error
			if err != nil {
				if errors.Is(err, database.ErrRecordNotFound) {
					err = d.cache.SetUsernamePlaceholder(ctx, username)
					if err != nil {
						logger.Warn("cache.SetUsernamePlaceholder error", logger.Err(err), logger.Any("username", username))
					}
					return nil, database.ErrRecordNotFound
				}
				return nil, err
			}
			// set cache
			err = d.cache.SetByUsername(ctx, username, table, 10*time.Minute)
			if err != nil {
				logger.Warn("cache.SetByUsername error", logger.Err(err), logger.Any("username", username))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.User)
		if !ok {
			return nil, database.ErrRecordNotFound
		}
		return table, nil
	}

	if d.cache.IsPlaceholderErr(err) {
		return nil, database.ErrRecordNotFound
	}

	return nil, err
}
