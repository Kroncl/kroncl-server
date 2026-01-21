package core

import (
	"fmt"
	"strings"
)

// only for PATCH queries

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

// SetBool добавляет bool (всегда)
func (u *Updater) SetBool(field string, value *bool) *Updater {
	if value != nil {
		return u.Set(field, *value)
	}
	return u
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

	query := fmt.Sprintf(
		"UPDATE %s SET %s, updated_at = NOW()",
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
