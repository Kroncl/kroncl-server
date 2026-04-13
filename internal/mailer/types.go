package mailer

// Recipient получатель письма
type Recipient struct {
	Email         string            `json:"email"`
	Substitutions map[string]string `json:"substitutions,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// Body части письма
type Body struct {
	HTML      string `json:"html,omitempty"`
	Plaintext string `json:"plaintext,omitempty"`
	AMP       string `json:"amp,omitempty"`
}

// Attachment вложение
type Attachment struct {
	Type    string `json:"type"`    // MIME type
	Name    string `json:"name"`    // имя файла
	Content string `json:"content"` // base64
}

// InlineAttachment inline вложение (для картинок)
type InlineAttachment struct {
	Type    string `json:"type"`
	Name    string `json:"name"`    // CID для ссылки в html
	Content string `json:"content"` // base64
}

// Options дополнительные опции
type Options struct {
	SendAt          string `json:"send_at,omitempty"`           // время отправки UTC
	UnsubscribeURL  string `json:"unsubscribe_url,omitempty"`   // кастомная ссылка отписки
	CustomBackendID int    `json:"custom_backend_id,omitempty"` // ID backend-домена
	SMTPPoolID      string `json:"smtp_pool_id,omitempty"`      // UUID пула SMTP
}

// Message полное сообщение
type Message struct {
	// Получатели (обязательно, макс 500)
	Recipients []Recipient `json:"recipients"`

	// Тема и тело (subject и хотя бы одно поле body обязательно)
	Subject string `json:"subject"`
	Body    Body   `json:"body"`

	// Отправитель
	FromEmail   string `json:"from_email,omitempty"`
	FromName    string `json:"from_name,omitempty"`
	ReplyTo     string `json:"reply_to,omitempty"`
	ReplyToName string `json:"reply_to_name,omitempty"`

	// Шаблоны
	TemplateID     string `json:"template_id,omitempty"`
	TemplateEngine string `json:"template_engine,omitempty"` // simple, velocity, liquid, none

	// Трекинг
	TrackLinks int `json:"track_links,omitempty"` // 1=вкл, 0=выкл
	TrackRead  int `json:"track_read,omitempty"`  // 1=вкл, 0=выкл

	// Теги и метаданные
	Tags           []string          `json:"tags,omitempty"` // макс 4 шт, до 50 символов
	GlobalSubs     map[string]string `json:"global_substitutions,omitempty"`
	GlobalMetadata map[string]string `json:"global_metadata,omitempty"`
	GlobalLanguage string            `json:"global_language,omitempty"` // be,de,en,es,fr,it,pl,pt,ru,ua,kz

	// Bypass опции (нужно разрешение поддержки)
	SkipUnsubscribe    int `json:"skip_unsubscribe,omitempty"`    // 1=пропустить блок отписки
	BypassGlobal       int `json:"bypass_global,omitempty"`       // игнорировать глобальные блокировки
	BypassUnavailable  int `json:"bypass_unavailable,omitempty"`  // игнорировать недоступные
	BypassUnsubscribed int `json:"bypass_unsubscribed,omitempty"` // игнорировать отписавшихся
	BypassComplained   int `json:"bypass_complained,omitempty"`   // игнорироваь пожаловавшихся

	// Дополнительно
	IdempotenceKey    string             `json:"idempotence_key,omitempty"` // для защиты от дублей
	Headers           map[string]string  `json:"headers,omitempty"`         // X-* заголовки
	Attachments       []Attachment       `json:"attachments,omitempty"`
	InlineAttachments []InlineAttachment `json:"inline_attachments,omitempty"`
	Options           *Options           `json:"options,omitempty"`
}

// SendRequest запрос к API
type SendRequest struct {
	Message Message `json:"message"`
}

// SendResponse успешный ответ
type SendResponse struct {
	Status       string            `json:"status"`
	JobID        string            `json:"job_id"`
	Emails       []string          `json:"emails,omitempty"`
	FailedEmails map[string]string `json:"failed_emails,omitempty"`
}

// ErrorResponse ответ с ошибкой
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
