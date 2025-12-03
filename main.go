package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"matrix-authorization-server/internal/accounts"
	"matrix-authorization-server/utils"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// грузим окружение
	err := godotenv.Load()
	if err != nil {
		log.Println("Внимание: не удалось загрузить .env файл, используем переменные окружения")
	}

	// мапим константы из конфига в dsn
	config := utils.LoadDBConfigFromEnv()
	dsn, err := utils.BuildDSN(config)
	if err != nil {
		log.Fatal("Ошибка формирования DSN:", err)
	}

	// инит пула к базе
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer pool.Close()

	// 1. Пользователь регистрируется
	email := "trtasdrtrt@example.com"
	name := "John Doe"
	password := "Password123"

	fmt.Println("🚀 Начинаем процесс регистрации...")
	fmt.Printf("📧 Email: %s\n", email)
	fmt.Printf("👤 Имя: %s\n", name)

	accountID, err := accounts.Create(pool, email, name, password)
	if err != nil {
		log.Fatal("❌ Ошибка регистрации:", err)
	}

	fmt.Printf("✅ Аккаунт создан! ID: %s\n", accountID)

	// 2. Генерируем и отправляем код подтверждения
	fmt.Println("\n📨 Генерируем код подтверждения...")
	code, err := accounts.GenerateConfirmationCode(pool, accountID, "email_confirmation", 6, 5)
	if err != nil {
		log.Fatal("❌ Ошибка генерации кода:", err)
	}

	// 3. Отправляем код на email
	fmt.Printf("✅ Код сгенерирован: %s\n", code)
	fmt.Println("📧 Отправляем код на email...")

	err = accounts.SendConfirmationEmail(email, code)
	if err != nil {
		log.Fatal("❌ Ошибка отправки email:", err)
	}

	fmt.Println("✅ Код отправлен на email!")
	fmt.Println("\n======================================")
	fmt.Printf("📋 Ваш код подтверждения: %s\n", code)
	fmt.Println("======================================")
	fmt.Println("\n💡 В реальности код был бы отправлен на указанный email")
	fmt.Println("   Сейчас вы можете ввести его вручную для тестирования")

	// 4. Запрашиваем ввод кода от пользователя
	reader := bufio.NewReader(os.Stdin)

	for attempt := 1; attempt <= 3; attempt++ {
		fmt.Printf("\n🔄 Попытка %d из 3\n", attempt)
		fmt.Print("➡️  Введите код подтверждения: ")

		userEnteredCode, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("❌ Ошибка чтения ввода:", err)
		}

		// Очищаем ввод от пробелов и переносов строк
		userEnteredCode = strings.TrimSpace(userEnteredCode)

		if userEnteredCode == "" {
			fmt.Println("⚠️  Код не может быть пустым. Попробуйте еще раз.")
			continue
		}

		// 5. Проверяем код
		fmt.Println("🔍 Проверяем код...")
		isValid, err := accounts.VerifyConfirmationCode(pool, accountID, userEnteredCode, "email_confirmation")
		if err != nil {
			log.Fatal("❌ Ошибка проверки кода:", err)
		}

		if isValid {
			// 6. Подтверждаем аккаунт
			fmt.Println("✅ Код верный! Подтверждаем аккаунт...")
			err = accounts.MarkAccountAsConfirmed(pool, accountID)
			if err != nil {
				log.Fatal("❌ Ошибка подтверждения аккаунта:", err)
			}
			fmt.Println("🎉 Email успешно подтвержден!")
			fmt.Println("✅ Регистрация завершена успешно!")

			// Показываем информацию о подтвержденном аккаунте
			showAccountInfo(pool, accountID)
			return
		} else {
			fmt.Println("❌ Неверный код.")

			// Проверяем, есть ли еще попытки
			if attempt < 3 {
				fmt.Println("💡 Попробуйте еще раз")
			} else {
				fmt.Println("🚫 Превышено количество попыток.")

				// Предлагаем отправить новый код
				fmt.Print("\n🔄 Отправить новый код? (y/n): ")
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))

				if response == "y" || response == "yes" || response == "да" {
					// Генерируем новый код
					newCode, err := accounts.GenerateConfirmationCode(pool, accountID, "email_confirmation", 6, 5)
					if err != nil {
						log.Fatal("❌ Ошибка генерации нового кода:", err)
					}

					fmt.Printf("\n✅ Новый код отправлен: %s\n", newCode)
					fmt.Println("📧 (В реальности был бы отправлен на email)")

					// Сбрасываем счетчик попыток
					attempt = 0
				} else {
					fmt.Println("👋 Выход из программы.")
					return
				}
			}
		}
	}
}

func showAccountInfo(pool *pgxpool.Pool, accountID string) {
	account, err := accounts.GetByID(pool, accountID)
	if err != nil {
		log.Println("⚠️  Не удалось получить информацию об аккаунте:", err)
		return
	}

	fmt.Println("\n📋 Информация об аккаунте:")
	fmt.Println("================================")
	fmt.Printf("🆔 ID: %s\n", account.Id)
	fmt.Printf("📧 Email: %s\n", account.Email)
	fmt.Printf("👤 Имя: %s\n", account.Name)
	fmt.Printf("🔑 Тип аутентификации: %s\n", account.AuthType)
	fmt.Printf("📊 Статус: %s\n", account.Status)
	fmt.Printf("📅 Создан: %s\n", account.CreatedAt)
	fmt.Printf("🔄 Обновлен: %s\n", account.UpdatedAt)
	fmt.Println("================================")
}
