package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
)

func initEmailI18n(t *testing.T) {
	t.Helper()
	require.NoError(t, i18n.Init())
}

func TestBuildEmailVerificationEmail(t *testing.T) {
	initEmailI18n(t)
	common.SystemName = "New API"
	common.Logo = ""

	tests := []struct {
		lang         string
		wantSubject  string
		wantTitle    string
		wantNotInSub string
	}{
		{i18n.LangZhCN, "【New API】邮箱验证码", "邮箱验证", ""},
		{i18n.LangZhTW, "【New API】信箱驗證碼", "信箱驗證", ""},
		{i18n.LangEn, "New API email verification code", "Email Verification", ""},
	}
	for _, tt := range tests {
		subject, content := BuildEmailVerificationEmail(tt.lang, "47f373")
		assert.Equal(t, tt.wantSubject, subject)
		assert.Contains(t, content, "47f373")
		assert.Contains(t, content, tt.wantTitle)
		assert.Contains(t, content, "<!DOCTYPE html>")
	}
}

func TestRenderEmailCardEscapesDynamicValues(t *testing.T) {
	initEmailI18n(t)
	common.SystemName = `<script>alert(1)</script>`
	common.Logo = ""

	_, content := BuildEmailVerificationEmail(i18n.LangEn, "42")
	assert.NotContains(t, content, "<script>alert(1)</script>")
	assert.Contains(t, content, "&lt;script&gt;")
	// The verification code itself must remain readable.
	assert.Contains(t, content, ">42</span>")
}

func TestRenderEmailCardLogoHandling(t *testing.T) {
	initEmailI18n(t)
	common.SystemName = "New API"

	common.Logo = "https://example.com/logo.png"
	content := renderEmailCard(i18n.LangEn, "Title", "<p>body</p>")
	assert.Contains(t, content, `<img src="https://example.com/logo.png"`)

	// Relative or non-HTTP logo values fall back to the text brand header.
	common.Logo = "/logo.png"
	content = renderEmailCard(i18n.LangEn, "Title", "<p>body</p>")
	assert.NotContains(t, content, "<img")
	assert.Contains(t, content, ">New API</div>")
}

func TestBuildPasswordResetEmail(t *testing.T) {
	initEmailI18n(t)
	common.SystemName = "New API"
	common.Logo = ""

	subject, content := BuildPasswordResetEmail(i18n.LangZhCN, "https://example.com/user/reset?email=a@b.com&token=xyz")
	require.True(t, strings.HasPrefix(subject, "【New API】"))
	assert.Contains(t, content, `href="https://example.com/user/reset?email=a@b.com&amp;token=xyz"`)
	assert.Contains(t, content, "重置密码")
}
