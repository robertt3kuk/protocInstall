package utils

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

// Unit тест для функции GetLastSegment.
func TestGetLastSegment(t *testing.T) {
	// Тест с нормальной строкой, содержащей несколько сегментов.
	ref := "http://example.com/path/to/resource"
	result := GetLastSegment(ref)
	assert.Equal(t, "resource", result, "Должен вернуть последний сегмент")

	// Тест с строкой, оканчивающейся на /
	ref = "http://example.com/path/to/resource/"
	result = GetLastSegment(ref)
	assert.Equal(t, "", result, "Должен вернуть пустую строку, если последний символ /")

	// Тест с одиночным сегментом.
	ref = "resource"
	result = GetLastSegment(ref)
	assert.Equal(t, "resource", result, "Должен вернуть саму строку, если нет /")

	// Тест с пустой строкой.
	ref = ""
	result = GetLastSegment(ref)
	assert.Equal(t, "", result, "Должен вернуть пустую строку при пустом вводе")

	// Тест с строкой, содержащей только один /
	ref = "/"
	result = GetLastSegment(ref)
	assert.Equal(t, "", result, "Должен вернуть пустую строку для '/'")
}

// Unit тест для GetSchemaName, когда схема присутствует.
func TestGetSchemaName_Success(t *testing.T) {
	// Создаем mock ссылку на схему с $ref.
	schemaRef := &openapi3.SchemaRef{
		Ref: "#/components/schemas/YourSchemaName",
	}

	// Проверяем успешное извлечение имени схемы.
	schemaName := GetSchemaName(schemaRef)
	assert.Equal(t, "YourSchemaName", schemaName, "Имя схемы должно быть 'YourSchemaName'")
}

// Unit тест для GetSchemaName, когда schemaRef равен nil.
func TestGetSchemaName_NilSchemaRef(t *testing.T) {
	// Передаем nil в качестве schemaRef.
	schemaName := GetSchemaName(nil)

	// Проверяем, что функция возвращает пустую строку.
	assert.Equal(t, "", schemaName, "При nil schemaRef функция должна вернуть пустую строку")
}

// Unit тест для GetSchemaName, когда Ref пустой.
func TestGetSchemaName_EmptyRef(t *testing.T) {
	// Создаем schemaRef с пустым Ref.
	schemaRef := &openapi3.SchemaRef{
		Ref: "",
	}

	// Проверяем, что функция возвращает пустую строку.
	schemaName := GetSchemaName(schemaRef)
	assert.Equal(t, "", schemaName, "При пустом Ref функция должна вернуть пустую строку")
}

// Unit тест для GetSchemaName, когда в Ref нет '/'.
func TestGetSchemaName_NoSlashInRef(t *testing.T) {
	// Создаем schemaRef без символов '/'.
	schemaRef := &openapi3.SchemaRef{
		Ref: "SimpleSchemaName",
	}

	// Проверяем, что функция возвращает саму строку Ref.
	schemaName := GetSchemaName(schemaRef)
	assert.Equal(t, "SimpleSchemaName", schemaName, "Функция должна вернуть сам Ref, если нет '/'")
}

// Unit тест для GetSchemaName, когда Ref заканчивается на '/'.
func TestGetSchemaName_EndsWithSlash(t *testing.T) {
	// Создаем schemaRef, который заканчивается на '/'.
	schemaRef := &openapi3.SchemaRef{
		Ref: "#/components/schemas/",
	}

	// Проверяем, что функция возвращает пустую строку.
	schemaName := GetSchemaName(schemaRef)
	assert.Equal(t, "", schemaName, "Функция должна вернуть пустую строку, если Ref заканчивается на '/'")
}
