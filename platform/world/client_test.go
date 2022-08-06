package world

import (
	"github.com/celskeggs/mediator/platform/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

type okExpect struct {
	Type  types.ProcArgumentAs
	Value string
}

func ok(t *testing.T, input string, command string, expected ...okExpect) {
	argTypes := make([]types.ProcArgumentAs, len(expected))
	for i, ex := range expected {
		argTypes[i] = ex.Type
	}
	parsed, err := parseVerb(input, argTypes)
	assert.NoError(t, err)
	assert.Equal(t, len(expected)+1, len(parsed), "while running: %s", input)
	assert.Equal(t, command, parsed[0])
	for i, ex := range expected {
		if i < len(parsed) {
			assert.Equal(t, ex.Value, parsed[i+1], "while running: %s", input)
		}
	}
}

func err(t *testing.T, input string, argTypes ...types.ProcArgumentAs) {
	_, err := parseVerb(input, argTypes)
	assert.Error(t, err)
}

func TestVerbParsing1(t *testing.T) {
	ok(t, "say", "say")
}

func TestVerbParsing2(t *testing.T) {
	ok(t, "say \"", "say",
		okExpect{types.ProcArgumentText, ""})
}

func TestVerbParsing3(t *testing.T) {
	ok(t, "say \"\"", "say",
		okExpect{types.ProcArgumentText, ""})
}

func TestVerbParsing4(t *testing.T) {
	ok(t, "say hello world", "say",
		okExpect{types.ProcArgumentText, "hello world"})
}

func TestVerbParsing5(t *testing.T) {
	err(t, "say hello world", types.ProcArgumentText, types.ProcArgumentText)
}

func TestVerbParsing6(t *testing.T) {
	err(t, "say \"hello\" world", types.ProcArgumentText)
}

func TestVerbParsing7(t *testing.T) {
	ok(t, "say \"hello world\" 123 123 123", "say",
		okExpect{types.ProcArgumentText, "hello world"},
		okExpect{types.ProcArgumentText, "123 123 123"})
}

func TestVerbParsing8(t *testing.T) {
	ok(t, "say \"abc\"abc\"", "say",
		okExpect{types.ProcArgumentText, "abc\"abc\""})
}

func TestVerbParsing9(t *testing.T) {
	ok(t, "say 1234", "say",
		okExpect{types.ProcArgumentText, "1234"})
}

func TestVerbParsing10(t *testing.T) {
	ok(t, "say \"1234", "say",
		okExpect{types.ProcArgumentText, "1234"})
}

func TestVerbParsing11(t *testing.T) {
	ok(t, "say \"hello world", "say",
		okExpect{types.ProcArgumentText, "hello world"})
}

func TestVerbParsing12(t *testing.T) {
	ok(t, "say \"hello world\"", "say",
		okExpect{types.ProcArgumentText, "hello world"})
}

func TestVerbParsing13(t *testing.T) {
	err(t, "say \"123\" \"abc\"", types.ProcArgumentText)
}

func TestVerbParsing14(t *testing.T) {
	ok(t, "say \"text abc 123 \"text\" abc 123\"", "say",
		okExpect{types.ProcArgumentText, "text abc 123 \"text\" abc 123"})
}

func TestVerbParsing15(t *testing.T) {
	ok(t, "say \"that's \"okay by me\" I guess\"", "say",
		okExpect{types.ProcArgumentText, "that's \"okay by me\" I guess"})
}

func TestVerbParsing16(t *testing.T) {
	err(t, "say \"that's okay by me\" Iguess\"", types.ProcArgumentText)
}

func TestVerbParsing17(t *testing.T) {
	ok(t, "say \"that's okay b\"y me\"", "say",
		okExpect{types.ProcArgumentText, "that's okay b\"y me\""})
}

func TestVerbParsing18(t *testing.T) {
	ok(t, "say \"that's okay b\"y me\" I guess", "say",
		okExpect{types.ProcArgumentText, "that's okay b\"y me\" I guess"})
}

func TestVerbParsing19(t *testing.T) {
	ok(t, "say \"that's okay b\"y me\" I guess\"", "say",
		okExpect{types.ProcArgumentText, "that's okay b\"y me\" I guess"})
}

func TestVerbParsing20(t *testing.T) {
	ok(t, "say \"that's okay b\"\"y me\" I guess", "say",
		okExpect{types.ProcArgumentText, "that's okay b\"\"y me\" I guess"})
}

func TestVerbParsing21(t *testing.T) {
	ok(t, "say \"that's okay b\"\"y me\" I guess\"", "say",
		okExpect{types.ProcArgumentText, "that's okay b\"\"y me\" I guess\""})
}

func TestVerbParsing22(t *testing.T) {
	ok(t, "say \"that's okay b\"\"y me\" I guess\" 123", "say",
		okExpect{types.ProcArgumentText, "that's okay b\"\"y me\" I guess\" 123"})
}

func TestVerbParsing23(t *testing.T) {
	ok(t, "say \"that's okay b\"\"y me\" I guess\" 123\"", "say",
		okExpect{types.ProcArgumentText, "that's okay b\"\"y me\" I guess\" 123"})
}

func TestVerbParsing24(t *testing.T) {
	err(t, "say \"that's okay b\"\"y me\" I guess\" 123\" 456", types.ProcArgumentText)
}

func TestVerbParsing25(t *testing.T) {
	err(t, "say \"that's okay b\"\"y me\" I guess\" 123\" 456\"", types.ProcArgumentText)
}

func TestVerbParsing26(t *testing.T) {
	ok(t, "say \"\"that's okay by me", "say",
		okExpect{types.ProcArgumentText, "\"that's okay by me"})
}

func TestVerbParsing27(t *testing.T) {
	ok(t, "say \"\"that's okay by me\"", "say",
		okExpect{types.ProcArgumentText, "\"that's okay by me\""})
}

func TestVerbParsing28(t *testing.T) {
	ok(t, "say \"\"that's okay by me\"\"", "say",
		okExpect{types.ProcArgumentText, "\"that's okay by me\"\""})
}

func TestVerbParsing29(t *testing.T) {
	ok(t, "say ok abc oeuthoenuthaeonthutoh aoeu 333", "say",
		okExpect{types.ProcArgumentText, "ok abc oeuthoenuthaeonthutoh aoeu 333"})
}

func TestVerbParsing30(t *testing.T) {
	ok(t, "say ok\"", "say",
		okExpect{types.ProcArgumentText, "ok\""})
}

func TestVerbParsing31(t *testing.T) {
	ok(t, "say 123 456 bye\"", "say",
		okExpect{types.ProcArgumentText, "123 456 bye\""})
}

func TestVerbParsing32(t *testing.T) {
	ok(t, "say 123 456 bye\" 123\"", "say",
		okExpect{types.ProcArgumentText, "123 456 bye\" 123\""})
}

func TestVerbParsing33(t *testing.T) {
	ok(t, "say hello w\"\"rld a\" b\"", "say",
		okExpect{types.ProcArgumentText, "hello w\"\"rld a\" b\""})
}

func TestVerbParsing34(t *testing.T) {
	ok(t, "say hello w\"\"rld\" a\" b\"", "say",
		okExpect{types.ProcArgumentText, "hello w\"\"rld\" a\" b\""})
}

func TestVerbParsing35(t *testing.T) {
	ok(t, "say hello w\"\"rld\" a\" b\" c\"", "say",
		okExpect{types.ProcArgumentText, "hello w\"\"rld\" a\" b\" c\""})
}

func TestVerbParsing36(t *testing.T) {
	ok(t, "say \"hello w\"\"rld a\" b\"", "say",
		okExpect{types.ProcArgumentText, "hello w\"\"rld a\" b\""})
}

func TestVerbParsing37(t *testing.T) {
	ok(t, "say \"hello w\"\"rld\" a\" b\"", "say",
		okExpect{types.ProcArgumentText, "hello w\"\"rld\" a\" b\""})
}

func TestVerbParsing38(t *testing.T) {
	err(t, "say \"hello w\"\"rld\" a\" b\" c\"", types.ProcArgumentText)
}

func TestVerbParsing39(t *testing.T) {
	ok(t, "say \"hello w\"\"rld a\" b\" c\"", "say",
		okExpect{types.ProcArgumentText, "hello w\"\"rld a\" b\" c"})
}

func TestVerbParsing40(t *testing.T) {
	err(t, "say \"hello w\"\"rld a\" b\" c\" d\"", 1)
}

func TestVerbParsing41(t *testing.T) {
	ok(t, "say hello\" world\"", "say",
		okExpect{types.ProcArgumentText, "hello\" world\""})
}

func TestVerbParsing42(t *testing.T) {
	ok(t, "say te\"st\" 123\"", "say",
		okExpect{types.ProcArgumentText, "te\"st\" 123\""})
}

func TestVerbParsing43(t *testing.T) {
	ok(t, "say te\"st\" 123\" 123\"", "say",
		okExpect{types.ProcArgumentText, "te\"st\" 123\" 123\""})
}

func TestVerbParsing44(t *testing.T) {
	ok(t, "say abc \"123\" def", "say",
		okExpect{types.ProcArgumentText, "abc \"123\" def"})
}

func TestVerbParsing45(t *testing.T) {
	ok(t, "say 123   456", "say",
		okExpect{types.ProcArgumentText, "123   456"})
}

func TestVerbParsing46(t *testing.T) {
	ok(t, "say \"123   456\"", "say",
		okExpect{types.ProcArgumentText, "123   456"})
}
