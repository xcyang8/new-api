package service

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/oauth"
	"github.com/QuantumNous/new-api/setting/system_setting"
)

func UnbindAccountOAuth(identity AuthIdentity, providerID int) error {
	enabled := model.AccountLoginMethods{
		Password: common.PasswordLoginEnabled,
		Passkey:  system_setting.PasskeySettingsSnapshot().Enabled,
		WeChat:   common.WeChatAuthEnabled,
	}
	for _, provider := range oauth.GetAllProviders() {
		if !provider.IsEnabled() {
			continue
		}
		if custom, ok := provider.(*oauth.GenericOAuthProvider); ok {
			enabled.CustomProviderIDs = append(enabled.CustomProviderIDs, custom.GetProviderId())
		} else {
			enabled.OAuthColumns = append(enabled.OAuthColumns, provider.ProviderUserIDColumn())
		}
	}
	return model.UnbindUserOAuthForSession(identity, providerID, enabled)
}

// NotifyAccountSecurityChange never includes credentials or tokens. The caller
// records delivery failure independently from the already-committed change.
// event must already be localized for lang (use i18n.T at the call site).
func NotifyAccountSecurityChange(lang, email, event string) error {
	if email == "" {
		return nil
	}
	args := map[string]any{"SystemName": common.SystemName}
	subject := i18n.Translate(lang, i18n.MsgEmailSecuritySubject, args)
	args["Event"] = event
	body := emailParagraph(i18n.Translate(lang, i18n.MsgEmailSecurityIntro, args)) +
		emailParagraph(i18n.Translate(lang, i18n.MsgEmailSecurityAdvice))
	content := renderEmailCard(lang, i18n.Translate(lang, i18n.MsgEmailSecurityTitle), body)
	return common.SendEmail(subject, email, content)
}
