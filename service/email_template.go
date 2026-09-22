package service

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
)

const emailFontStack = "-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif"

// renderEmailCard wraps a localized title and body fragment in the branded card
// layout shared by all system emails. bodyHTML must already be built from
// escaped values; title is escaped here.
func renderEmailCard(lang, title, bodyHTML string) string {
	systemName := html.EscapeString(common.SystemName)
	header := fmt.Sprintf(`<div style="font-size:20px;font-weight:600;color:#111827;text-align:center;font-family:%s;">%s</div>`, emailFontStack, systemName)
	if logo := strings.TrimSpace(common.Logo); strings.HasPrefix(logo, "http://") || strings.HasPrefix(logo, "https://") {
		header = fmt.Sprintf(`<img src="%s" alt="%s" style="display:block;margin:0 auto;max-height:48px;max-width:200px;">`, html.EscapeString(logo), systemName)
	}
	footerBrand := systemName
	if serverAddress := strings.TrimRight(strings.TrimSpace(system_setting.ServerAddress), "/"); serverAddress != "" {
		footerBrand = fmt.Sprintf(`<a href="%s" style="color:#9ca3af;text-decoration:none;">%s</a>`, html.EscapeString(serverAddress), systemName)
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;background-color:#f3f4f6;">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f3f4f6;padding:32px 16px;">
<tr><td align="center">
<table role="presentation" width="560" cellpadding="0" cellspacing="0" style="max-width:560px;width:100%%;background-color:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 1px 3px rgba(0,0,0,0.08);">
<tr><td style="padding:32px 40px 8px;">%s</td></tr>
<tr><td style="padding:16px 40px 0;font-size:22px;font-weight:600;color:#111827;text-align:center;font-family:%s;">%s</td></tr>
<tr><td style="padding:24px 40px 32px;font-size:14px;line-height:1.7;color:#374151;font-family:%s;">%s</td></tr>
</table>
<table role="presentation" width="560" cellpadding="0" cellspacing="0" style="max-width:560px;width:100%%;">
<tr><td style="padding:24px 40px;text-align:center;font-size:12px;line-height:1.6;color:#9ca3af;font-family:%s;">
<p style="margin:0 0 8px;">%s</p>
<p style="margin:0;">&copy; %d %s</p>
</td></tr>
</table>
</td></tr>
</table>
</body>
</html>`, html.EscapeString(lang), header, emailFontStack, html.EscapeString(title), emailFontStack, bodyHTML,
		emailFontStack, html.EscapeString(i18n.Translate(lang, i18n.MsgEmailFooterAutoSent)), time.Now().Year(), footerBrand)
}

// emailParagraph renders an escaped plain-text paragraph.
func emailParagraph(text string) string {
	return fmt.Sprintf(`<p style="margin:0 0 12px;">%s</p>`, html.EscapeString(text))
}

// emailCodeBlock renders a verification code as a large mono-spaced badge.
func emailCodeBlock(code string) string {
	return fmt.Sprintf(`<div style="margin:24px 0;text-align:center;"><span style="display:inline-block;padding:14px 32px;font-size:32px;font-weight:700;letter-spacing:6px;font-family:SFMono-Regular,Menlo,Consolas,'Courier New',monospace;color:#111827;background-color:#f3f4f6;border-radius:8px;">%s</span></div>`, html.EscapeString(code))
}

func emailCodeParagraphs(lang string, minutes int) string {
	return emailParagraph(i18n.Translate(lang, i18n.MsgEmailCodeExpiry, map[string]any{"Minutes": minutes})) +
		emailParagraph(i18n.Translate(lang, i18n.MsgEmailCodeIgnore))
}

// BuildEmailVerificationEmail builds the registration verification-code email.
func BuildEmailVerificationEmail(lang, code string) (subject, content string) {
	args := map[string]any{"SystemName": common.SystemName}
	subject = i18n.Translate(lang, i18n.MsgEmailVerificationSubject, args)
	body := emailParagraph(i18n.Translate(lang, i18n.MsgEmailVerificationIntro, args)) +
		emailCodeBlock(code) +
		emailCodeParagraphs(lang, common.VerificationValidMinutes)
	return subject, renderEmailCard(lang, i18n.Translate(lang, i18n.MsgEmailVerificationTitle), body)
}

// BuildPasswordResetEmail builds the password reset email with a button link.
func BuildPasswordResetEmail(lang, link string) (subject, content string) {
	args := map[string]any{"SystemName": common.SystemName}
	subject = i18n.Translate(lang, i18n.MsgEmailPasswordResetSubject, args)
	escapedLink := html.EscapeString(link)
	button := fmt.Sprintf(`<div style="margin:24px 0;text-align:center;"><a href="%s" style="display:inline-block;padding:12px 36px;background-color:#2563eb;color:#ffffff;text-decoration:none;border-radius:8px;font-size:15px;font-weight:600;">%s</a></div>`,
		escapedLink, html.EscapeString(i18n.Translate(lang, i18n.MsgEmailPasswordResetButton)))
	linkFallback := emailParagraph(i18n.Translate(lang, i18n.MsgEmailPasswordResetLinkFallback)) +
		fmt.Sprintf(`<p style="margin:0 0 12px;word-break:break-all;font-size:13px;"><a href="%s" style="color:#2563eb;">%s</a></p>`, escapedLink, escapedLink)
	body := emailParagraph(i18n.Translate(lang, i18n.MsgEmailPasswordResetIntro, args)) +
		button + linkFallback +
		emailParagraph(i18n.Translate(lang, i18n.MsgEmailPasswordResetExpiry, map[string]any{"Minutes": common.VerificationValidMinutes})) +
		emailParagraph(i18n.Translate(lang, i18n.MsgEmailPasswordResetIgnore))
	return subject, renderEmailCard(lang, i18n.Translate(lang, i18n.MsgEmailPasswordResetTitle), body)
}

// buildEmailBindingEmail builds one email-binding confirmation email. introKey
// selects the new-address or current-address wording.
func buildEmailBindingEmail(lang, introKey, code string) (subject, content string) {
	args := map[string]any{"SystemName": common.SystemName}
	subject = i18n.Translate(lang, i18n.MsgEmailBindingSubject, args)
	body := emailParagraph(i18n.Translate(lang, introKey, args)) +
		emailCodeBlock(code) +
		emailCodeParagraphs(lang, int(model.EmailBindingTTL/time.Minute))
	return subject, renderEmailCard(lang, i18n.Translate(lang, i18n.MsgEmailBindingTitle), body)
}
