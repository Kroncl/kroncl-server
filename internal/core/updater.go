package core

import (
	"encoding/json"
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

// GetSets возвращает количество полей для обновления
func (u *Updater) GetSets() int {
	return len(u.sets)
}

// HasChanges возвращает true если есть что обновлять
func (u *Updater) HasChanges() bool {
	return len(u.sets) > 0
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

// SetNull устанавливает NULL для поля
func (u *Updater) SetNull(field string) *Updater {
	u.sets = append(u.sets, fmt.Sprintf("%s = NULL", field))
	return u
}

// SetNullableString добавляет строку или NULL
func (u *Updater) SetNullableString(field string, value *string) *Updater {
	if value != nil {
		return u.Set(field, *value)
	}
	return u.SetNull(field)
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

// SetJSONB добавляет JSONB поле (принимает map или struct и сериализует в JSON)
func (u *Updater) SetJSONB(field string, value interface{}) *Updater {
	if value != nil {
		// Проверяем, может быть value уже []byte?
		switch v := value.(type) {
		case []byte:
			// Если это уже байты, проверяем что это валидный JSON
			if len(v) > 0 {
				u.sets = append(u.sets, fmt.Sprintf("%s = $%d", field, len(u.args)+1))
				u.args = append(u.args, v)
			}
		default:
			// Сериализуем в JSON
			jsonBytes, err := json.Marshal(v)
			if err == nil && len(jsonBytes) > 0 {
				u.sets = append(u.sets, fmt.Sprintf("%s = $%d", field, len(u.args)+1))
				u.args = append(u.args, jsonBytes)
			}
			// Если ошибка маршалинга - просто игнорируем (можно логировать)
		}
	}
	return u
}

// SetJSONBIfNotNil добавляет JSONB поле только если value не nil
func (u *Updater) SetJSONBIfNotNil(field string, value *map[string]interface{}) *Updater {
	if value != nil {
		return u.SetJSONB(field, *value)
	}
	return u
}

// SetJSONBNull устанавливает NULL для JSONB поля
func (u *Updater) SetJSONBNull(field string) *Updater {
	u.sets = append(u.sets, fmt.Sprintf("%s = NULL", field))
	return u
}

func NullIfEmptyPtr(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
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
			newPlaceholder := fmt.Sprintf("$%d", len(u.args)+i+1)
			whereWithParams = strings.Replace(whereWithParams, oldPlaceholder, newPlaceholder, 1)
		}
		// Добавляем аргументы WHERE
		u.args = append(u.args, u.whereArgs...)
		query += " WHERE " + whereWithParams
	}

	return query, u.args
}

// NullIfEmpty возвращает nil если строка пустая, иначе указатель на строку
func NullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// JSONBIfEmpty возвращает nil если map пустой, иначе map
func JSONBIfEmpty(m map[string]interface{}) *map[string]interface{} {
	if len(m) == 0 {
		return nil
	}
	return &m
}
