// package core/updater.go
package core

import (
	"fmt"
	"strings"
	"time"
)

type Updater struct {
	table     string
	sets      []string
	args      []interface{}
	where     string
	whereArgs []interface{}
}

func NewUpdater(table string) *Updater {
	return &Updater{
		table: table,
		args:  make([]interface{}, 0),
		sets:  make([]string, 0),
	}
}

// Set добавляет поле для обновления (если не nil)
func (u *Updater) Set(field string, value interface{}) *Updater {
	if value != nil {
		u.sets = append(u.sets, fmt.Sprintf("%s = $%d", field, len(u.args)+1))
		u.args = append(u.args, value)
	}
	return u
}

// SetString добавляет строку (если не пустая)
func (u *Updater) SetString(field, value string) *Updater {
	if value != "" {
		return u.Set(field, value)
	}
	return u
}

// SetNullableString добавляет строку или NULL
func (u *Updater) SetNullableString(field string, value *string) *Updater {
	return u.Set(field, value)
}

// SetTime добавляет время
func (u *Updater) SetTime(field string, value time.Time) *Updater {
	return u.Set(field, value)
}

// SetBool добавляет bool
func (u *Updater) SetBool(field string, value bool) *Updater {
	return u.Set(field, value)
}

// SetInt добавляет int
func (u *Updater) SetInt(field string, value int) *Updater {
	return u.Set(field, value)
}

// SetFloat добавляет float
func (u *Updater) SetFloat(field string, value float64) *Updater {
	return u.Set(field, value)
}

// Where добавляет условие
func (u *Updater) Where(condition string, args ...interface{}) *Updater {
	u.where = condition
	u.whereArgs = args
	return u
}

// Build возвращает SQL и аргументы
func (u *Updater) Build() (string, []interface{}) {
	if len(u.sets) == 0 {
		return "", nil
	}

	// Всегда добавляем updated_at
	u.sets = append(u.sets, fmt.Sprintf("updated_at = $%d", len(u.args)+1))
	u.args = append(u.args, time.Now())

	query := fmt.Sprintf(
		"UPDATE %s SET %s",
		u.table,
		strings.Join(u.sets, ", "),
	)

	// Добавляем WHERE
	if u.where != "" {
		// Корректируем номера параметров в WHERE
		whereWithParams := u.where
		for i := range u.whereArgs {
			oldPlaceholder := fmt.Sprintf("$%d", i+1)
			newPlaceholder := fmt.Sprintf("$%d", len(u.args)+1)
			whereWithParams = strings.Replace(whereWithParams, oldPlaceholder, newPlaceholder, 1)
			u.args = append(u.args, u.whereArgs[i])
		}
		query += " WHERE " + whereWithParams
	}

	return query, u.args
}

func NullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
