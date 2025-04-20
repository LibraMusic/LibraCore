package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/amazon"
	"github.com/markbates/goth/providers/apple"
	"github.com/markbates/goth/providers/auth0"
	"github.com/markbates/goth/providers/azureadv2"
	"github.com/markbates/goth/providers/battlenet"
	"github.com/markbates/goth/providers/bitbucket"
	"github.com/markbates/goth/providers/bitly"
	"github.com/markbates/goth/providers/box"
	"github.com/markbates/goth/providers/classlink"
	"github.com/markbates/goth/providers/cloudfoundry"
	"github.com/markbates/goth/providers/cognito"
	"github.com/markbates/goth/providers/dailymotion"
	"github.com/markbates/goth/providers/deezer"
	"github.com/markbates/goth/providers/digitalocean"
	"github.com/markbates/goth/providers/discord"
	"github.com/markbates/goth/providers/dropbox"
	"github.com/markbates/goth/providers/eveonline"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/fitbit"
	"github.com/markbates/goth/providers/gitea"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/heroku"
	"github.com/markbates/goth/providers/hubspot"
	"github.com/markbates/goth/providers/influxcloud"
	"github.com/markbates/goth/providers/instagram"
	"github.com/markbates/goth/providers/intercom"
	"github.com/markbates/goth/providers/kakao"
	"github.com/markbates/goth/providers/lark"
	"github.com/markbates/goth/providers/lastfm"
	"github.com/markbates/goth/providers/line"
	"github.com/markbates/goth/providers/linkedin"
	"github.com/markbates/goth/providers/mastodon"
	"github.com/markbates/goth/providers/meetup"
	"github.com/markbates/goth/providers/microsoftonline"
	"github.com/markbates/goth/providers/naver"
	"github.com/markbates/goth/providers/nextcloud"
	"github.com/markbates/goth/providers/okta"
	"github.com/markbates/goth/providers/onedrive"
	"github.com/markbates/goth/providers/openidConnect"
	"github.com/markbates/goth/providers/oura"
	"github.com/markbates/goth/providers/patreon"
	"github.com/markbates/goth/providers/paypal"
	"github.com/markbates/goth/providers/salesforce"
	"github.com/markbates/goth/providers/seatalk"
	"github.com/markbates/goth/providers/shopify"
	"github.com/markbates/goth/providers/slack"
	"github.com/markbates/goth/providers/soundcloud"
	"github.com/markbates/goth/providers/spotify"
	"github.com/markbates/goth/providers/steam"
	"github.com/markbates/goth/providers/strava"
	"github.com/markbates/goth/providers/stripe"
	"github.com/markbates/goth/providers/tiktok"
	"github.com/markbates/goth/providers/tumblr"
	"github.com/markbates/goth/providers/twitch"
	"github.com/markbates/goth/providers/twitterv2"
	"github.com/markbates/goth/providers/uber"
	"github.com/markbates/goth/providers/vk"
	"github.com/markbates/goth/providers/wechat"
	"github.com/markbates/goth/providers/wecom"
	"github.com/markbates/goth/providers/wepay"
	"github.com/markbates/goth/providers/xero"
	"github.com/markbates/goth/providers/yahoo"
	"github.com/markbates/goth/providers/yammer"
	"github.com/markbates/goth/providers/yandex"
	"github.com/markbates/goth/providers/zoom"
)

type OAuthProvider struct {
	ID                string            `yaml:"id"`
	Key               string            `yaml:"key,alias=client_id"`
	Secret            string            `yaml:"secret,alias=client_secret"`
	AdditionalDetails map[string]string `yaml:",inline"`
}

func (p OAuthProvider) GetName() string {
	return providerNames[p.ID]
}

func (p OAuthProvider) GetProvider() (goth.Provider, error) {
	provider, err := goth.GetProvider(p.ID)
	if err != nil {
		factory, exists := providerFactories[p.ID]
		if !exists {
			return nil, fmt.Errorf("unsupported provider: %s", p.ID)
		}
		return factory(p)
	}
	return provider, nil
}

type providerFactory func(OAuthProvider) (goth.Provider, error)

var providerFactories = map[string]providerFactory{}

func RegisterProvider(id string, factory providerFactory) {
	providerFactories[id] = factory
}

//nolint:gocyclo,funlen // This function is long but we need a way to register all the providers
func RegisterBuiltInProviders(publicURL string) {
	RegisterProvider("amazon", func(p OAuthProvider) (goth.Provider, error) {
		return amazon.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/amazon/callback", publicURL)), nil
	})
	RegisterProvider("apple", func(p OAuthProvider) (goth.Provider, error) {
		return apple.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/auth/apple/callback", publicURL),
			nil,
			apple.ScopeName,
			apple.ScopeEmail,
		), nil
	})
	RegisterProvider("auth0", func(p OAuthProvider) (goth.Provider, error) {
		domain := p.AdditionalDetails["domain"]
		if domain == "" {
			return nil, errors.New("missing domain for auth0 provider")
		}
		return auth0.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/auth0/callback", publicURL), domain), nil
	})
	RegisterProvider("azuread", func(p OAuthProvider) (goth.Provider, error) {
		return azureadv2.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/auth/azuread/callback", publicURL),
			azureadv2.ProviderOptions{},
		), nil
	})
	RegisterProvider("battlenet", func(p OAuthProvider) (goth.Provider, error) {
		return battlenet.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/battlenet/callback", publicURL)), nil
	})
	RegisterProvider("bitbucket", func(p OAuthProvider) (goth.Provider, error) {
		return bitbucket.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/bitbucket/callback", publicURL)), nil
	})
	RegisterProvider("bitly", func(p OAuthProvider) (goth.Provider, error) {
		return bitly.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/bitly/callback", publicURL)), nil
	})
	RegisterProvider("box", func(p OAuthProvider) (goth.Provider, error) {
		return box.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/box/callback", publicURL)), nil
	})
	RegisterProvider("classlink", func(p OAuthProvider) (goth.Provider, error) {
		return classlink.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/classlink/callback", publicURL)), nil
	})
	RegisterProvider("cloudfoundry", func(p OAuthProvider) (goth.Provider, error) {
		authURL := p.AdditionalDetails["auth_url"]
		if authURL == "" {
			return nil, errors.New("missing auth_url for cloudfoundry provider")
		}
		return cloudfoundry.New(authURL, p.Key, p.Secret, fmt.Sprintf("%s/auth/cloudfoundry/callback", publicURL)), nil
	})
	RegisterProvider("cognito", func(p OAuthProvider) (goth.Provider, error) {
		baseURL := p.AdditionalDetails["base_url"]
		if baseURL != "" {
			return cognito.New(p.Key, p.Secret, baseURL, fmt.Sprintf("%s/auth/cognito/callback", publicURL)), nil
		}
		authURL := p.AdditionalDetails["auth_url"]
		tokenURL := p.AdditionalDetails["token_url"]
		issuerURL := p.AdditionalDetails["issuer_url"]
		profileURL := p.AdditionalDetails["profile_url"]
		if authURL == "" || tokenURL == "" || issuerURL == "" || profileURL == "" {
			return nil, errors.New(
				"missing base_url or auth_url, token_url, issuer_url, or profile_url for cognito provider",
			)
		}
		return cognito.NewCustomisedURL(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/auth/cognito/callback", publicURL),
			authURL,
			tokenURL,
			issuerURL,
			profileURL,
		), nil
	})
	RegisterProvider("dailymotion", func(p OAuthProvider) (goth.Provider, error) {
		return dailymotion.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/dailymotion/callback", publicURL), "email"), nil
	})
	RegisterProvider("deezer", func(p OAuthProvider) (goth.Provider, error) {
		return deezer.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/deezer/callback", publicURL), "email"), nil
	})
	RegisterProvider("digitalocean", func(p OAuthProvider) (goth.Provider, error) {
		return digitalocean.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/digitalocean/callback", publicURL), "read"), nil
	})
	RegisterProvider("discord", func(p OAuthProvider) (goth.Provider, error) {
		return discord.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/auth/discord/callback", publicURL),
			discord.ScopeIdentify,
			discord.ScopeEmail,
		), nil
	})
	RegisterProvider("dropbox", func(p OAuthProvider) (goth.Provider, error) {
		return dropbox.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/dropbox/callback", publicURL)), nil
	})
	RegisterProvider("eveonline", func(p OAuthProvider) (goth.Provider, error) {
		return eveonline.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/eveonline/callback", publicURL)), nil
	})
	RegisterProvider("facebook", func(p OAuthProvider) (goth.Provider, error) {
		return facebook.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/facebook/callback", publicURL)), nil
	})
	RegisterProvider("fitbit", func(p OAuthProvider) (goth.Provider, error) {
		return fitbit.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/fitbit/callback", publicURL)), nil
	})
	RegisterProvider("gitea", func(p OAuthProvider) (goth.Provider, error) {
		authURL := p.AdditionalDetails["auth_url"]
		tokenURL := p.AdditionalDetails["token_url"]
		profileURL := p.AdditionalDetails["profile_url"]
		if authURL != "" || tokenURL != "" || profileURL != "" {
			if authURL == "" || tokenURL == "" || profileURL == "" {
				return nil, errors.New("missing auth_url, token_url, or profile_url for gitea provider")
			}
			return gitea.NewCustomisedURL(
				p.Key,
				p.Secret,
				fmt.Sprintf("%s/auth/gitea/callback", publicURL),
				authURL,
				tokenURL,
				profileURL,
			), nil
		}
		return gitea.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/gitea/callback", publicURL)), nil
	})
	RegisterProvider("github", func(p OAuthProvider) (goth.Provider, error) {
		authURL := p.AdditionalDetails["auth_url"]
		tokenURL := p.AdditionalDetails["token_url"]
		profileURL := p.AdditionalDetails["profile_url"]
		emailURL := p.AdditionalDetails["email_url"]
		if authURL != "" || tokenURL != "" || profileURL != "" || emailURL != "" {
			if authURL == "" || tokenURL == "" || profileURL == "" || emailURL == "" {
				return nil, errors.New("missing auth_url, token_url, profile_url, or email_url for github provider")
			}
			return github.NewCustomisedURL(
				p.Key,
				p.Secret,
				fmt.Sprintf("%s/auth/github/callback", publicURL),
				authURL,
				tokenURL,
				profileURL,
				emailURL,
			), nil
		}
		return github.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/github/callback", publicURL)), nil
	})
	RegisterProvider("gitlab", func(p OAuthProvider) (goth.Provider, error) {
		authURL := p.AdditionalDetails["auth_url"]
		tokenURL := p.AdditionalDetails["token_url"]
		profileURL := p.AdditionalDetails["profile_url"]
		if authURL != "" || tokenURL != "" || profileURL != "" {
			if authURL == "" || tokenURL == "" || profileURL == "" {
				return nil, errors.New("missing auth_url, token_url, or profile_url for gitlab provider")
			}
			return gitlab.NewCustomisedURL(
				p.Key,
				p.Secret,
				fmt.Sprintf("%s/auth/gitlab/callback", publicURL),
				authURL,
				tokenURL,
				profileURL,
			), nil
		}
		return gitlab.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/gitlab/callback", publicURL)), nil
	})
	RegisterProvider("google", func(p OAuthProvider) (goth.Provider, error) {
		return google.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/google/callback", publicURL)), nil
	})
	RegisterProvider("heroku", func(p OAuthProvider) (goth.Provider, error) {
		return heroku.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/heroku/callback", publicURL)), nil
	})
	RegisterProvider("hubspot", func(p OAuthProvider) (goth.Provider, error) {
		return hubspot.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/hubspot/callback", publicURL)), nil
	})
	RegisterProvider("influxcloud", func(p OAuthProvider) (goth.Provider, error) {
		authURL := p.AdditionalDetails["auth_url"]
		tokenURL := p.AdditionalDetails["token_url"]
		userAPIEndpoint := p.AdditionalDetails["user_api_endpoint"]
		if authURL != "" || tokenURL != "" || userAPIEndpoint != "" {
			if authURL == "" || tokenURL == "" || userAPIEndpoint == "" {
				return nil, errors.New("missing auth_url, token_url, or user_api_endpoint for influxcloud provider")
			}
			return influxcloud.NewCustomisedURL(
				p.Key,
				p.Secret,
				fmt.Sprintf("%s/auth/influxcloud/callback", publicURL),
				authURL,
				tokenURL,
				userAPIEndpoint,
				"userscope",
			), nil
		}
		return influxcloud.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/auth/influxcloud/callback", publicURL),
			"userscope",
		), nil
	})
	RegisterProvider("instagram", func(p OAuthProvider) (goth.Provider, error) {
		return instagram.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/instagram/callback", publicURL)), nil
	})
	RegisterProvider("intercom", func(p OAuthProvider) (goth.Provider, error) {
		return intercom.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/intercom/callback", publicURL)), nil
	})
	RegisterProvider("kakao", func(p OAuthProvider) (goth.Provider, error) {
		return kakao.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/kakao/callback", publicURL)), nil
	})
	RegisterProvider("lark", func(p OAuthProvider) (goth.Provider, error) {
		return lark.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/lark/callback", publicURL)), nil
	})
	RegisterProvider("lastfm", func(p OAuthProvider) (goth.Provider, error) {
		return lastfm.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/lastfm/callback", publicURL)), nil
	})
	RegisterProvider("line", func(p OAuthProvider) (goth.Provider, error) {
		return line.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/auth/line/callback", publicURL),
			"profile",
			"openid",
			"email",
		), nil
	})
	RegisterProvider("linkedin", func(p OAuthProvider) (goth.Provider, error) {
		return linkedin.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/linkedin/callback", publicURL)), nil
	})
	RegisterProvider("mastodon", func(p OAuthProvider) (goth.Provider, error) {
		instanceURL := p.AdditionalDetails["instance_url"]
		if instanceURL != "" {
			return mastodon.NewCustomisedURL(
				p.Key,
				p.Secret,
				fmt.Sprintf("%s/auth/mastodon/callback", publicURL),
				instanceURL,
				"read:accounts",
			), nil
		}
		return mastodon.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/mastodon/callback", publicURL), "read:accounts"), nil
	})
	RegisterProvider("meetup", func(p OAuthProvider) (goth.Provider, error) {
		return meetup.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/meetup/callback", publicURL)), nil
	})
	RegisterProvider("microsoftonline", func(p OAuthProvider) (goth.Provider, error) {
		return microsoftonline.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/microsoftonline/callback", publicURL)), nil
	})
	RegisterProvider("naver", func(p OAuthProvider) (goth.Provider, error) {
		return naver.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/naver/callback", publicURL)), nil
	})
	RegisterProvider("nextcloud", func(p OAuthProvider) (goth.Provider, error) {
		nextcloudURL := p.AdditionalDetails["nextcloud_url"]
		if nextcloudURL != "" {
			return nextcloud.NewCustomisedDNS(
				p.Key,
				p.Secret,
				fmt.Sprintf("%s/auth/nextcloud/callback", publicURL),
				nextcloudURL,
			), nil
		}
		authURL := p.AdditionalDetails["auth_url"]
		tokenURL := p.AdditionalDetails["token_url"]
		profileURL := p.AdditionalDetails["profile_url"]
		if authURL == "" || tokenURL == "" || profileURL == "" {
			return nil, errors.New(
				"missing nextcloud_url or auth_url, token_url, or profile_url for nextcloud provider",
			)
		}
		return nextcloud.NewCustomisedURL(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/auth/nextcloud/callback", publicURL),
			authURL,
			tokenURL,
			profileURL,
		), nil
	})
	RegisterProvider("okta", func(p OAuthProvider) (goth.Provider, error) {
		orgURL := p.AdditionalDetails["org_url"]
		if orgURL != "" {
			return okta.New(p.Key, p.Secret, orgURL, fmt.Sprintf("%s/auth/okta/callback", publicURL)), nil
		}
		authURL := p.AdditionalDetails["auth_url"]
		tokenURL := p.AdditionalDetails["token_url"]
		issuerURL := p.AdditionalDetails["issuer_url"]
		profileURL := p.AdditionalDetails["profile_url"]
		if authURL == "" || tokenURL == "" || issuerURL == "" || profileURL == "" {
			return nil, errors.New(
				"missing org_url or auth_url, token_url, issuer_url, or profile_url for okta provider",
			)
		}
		return okta.NewCustomisedURL(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/auth/okta/callback", publicURL),
			authURL,
			tokenURL,
			issuerURL,
			profileURL,
		), nil
	})
	RegisterProvider("onedrive", func(p OAuthProvider) (goth.Provider, error) {
		return onedrive.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/onedrive/callback", publicURL)), nil
	})
	RegisterProvider("openid-connect", func(p OAuthProvider) (goth.Provider, error) {
		name := p.AdditionalDetails["name"]
		realName := "openid-connect"
		if name != "" {
			realName = strings.ToLower(name) + "-oidc"
		}

		autoDiscoveryURL := p.AdditionalDetails["auto_discovery_url"]
		if autoDiscoveryURL != "" {
			return openidConnect.NewNamed(
				name,
				p.Key,
				p.Secret,
				fmt.Sprintf("%s/auth/%s/callback", publicURL, realName),
				autoDiscoveryURL,
			)
		}
		authURL := p.AdditionalDetails["auth_url"]
		tokenURL := p.AdditionalDetails["token_url"]
		issuerURL := p.AdditionalDetails["issuer_url"]
		userInfoURL := p.AdditionalDetails["user_info_url"]
		endSessionEndpoint := p.AdditionalDetails["end_session_endpoint"]
		if authURL == "" || tokenURL == "" || issuerURL == "" || userInfoURL == "" || endSessionEndpoint == "" {
			return nil, fmt.Errorf(
				"missing auto_discovery_url or auth_url, token_url, issuer_url, user_info_url, or end_session_endpoint for openid-connect provider with name %s",
				name,
			)
		}
		provider, err := openidConnect.NewCustomisedURL(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/auth/%s/callback", publicURL, realName),
			authURL,
			tokenURL,
			issuerURL,
			userInfoURL,
			endSessionEndpoint,
		)
		if provider != nil {
			provider.SetName(realName)
		}
		return provider, err
	})
	RegisterProvider("oura", func(p OAuthProvider) (goth.Provider, error) {
		return oura.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/auth/oura/callback", publicURL),
			oura.ScopeEmail,
			oura.ScopePersonal,
		), nil
	})
	RegisterProvider("patreon", func(p OAuthProvider) (goth.Provider, error) {
		authURL := p.AdditionalDetails["auth_url"]
		tokenURL := p.AdditionalDetails["token_url"]
		profileURL := p.AdditionalDetails["profile_url"]
		if authURL != "" || tokenURL != "" || profileURL != "" {
			if authURL == "" || tokenURL == "" || profileURL == "" {
				return nil, errors.New("missing auth_url, token_url, or profile_url for patreon provider")
			}
			return patreon.NewCustomisedURL(
				p.Key,
				p.Secret,
				fmt.Sprintf("%s/auth/patreon/callback", publicURL),
				authURL,
				tokenURL,
				profileURL,
			), nil
		}
		return patreon.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/patreon/callback", publicURL)), nil
	})
	RegisterProvider("paypal", func(p OAuthProvider) (goth.Provider, error) {
		authURL := p.AdditionalDetails["auth_url"]
		tokenURL := p.AdditionalDetails["token_url"]
		profileURL := p.AdditionalDetails["profile_url"]
		if authURL != "" || tokenURL != "" || profileURL != "" {
			if authURL == "" || tokenURL == "" || profileURL == "" {
				return nil, errors.New("missing auth_url, token_url, or profile_url for paypal provider")
			}
			return paypal.NewCustomisedURL(
				p.Key,
				p.Secret,
				fmt.Sprintf("%s/auth/paypal/callback", publicURL),
				authURL,
				tokenURL,
				profileURL,
			), nil
		}
		return paypal.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/paypal/callback", publicURL)), nil
	})
	RegisterProvider("salesforce", func(p OAuthProvider) (goth.Provider, error) {
		return salesforce.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/salesforce/callback", publicURL)), nil
	})
	RegisterProvider("seatalk", func(p OAuthProvider) (goth.Provider, error) {
		return seatalk.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/seatalk/callback", publicURL)), nil
	})
	RegisterProvider("shopify", func(p OAuthProvider) (goth.Provider, error) {
		return shopify.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/auth/shopify/callback", publicURL),
			shopify.ScopeReadCustomers,
		), nil
	})
	RegisterProvider("slack", func(p OAuthProvider) (goth.Provider, error) {
		return slack.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/slack/callback", publicURL)), nil
	})
	RegisterProvider("soundcloud", func(p OAuthProvider) (goth.Provider, error) {
		return soundcloud.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/soundcloud/callback", publicURL)), nil
	})
	RegisterProvider("spotify", func(p OAuthProvider) (goth.Provider, error) {
		return spotify.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/spotify/callback", publicURL)), nil
	})
	RegisterProvider("steam", func(p OAuthProvider) (goth.Provider, error) {
		return steam.New(p.Key, fmt.Sprintf("%s/auth/steam/callback", publicURL)), nil
	})
	RegisterProvider("strava", func(p OAuthProvider) (goth.Provider, error) {
		return strava.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/strava/callback", publicURL)), nil
	})
	RegisterProvider("stripe", func(p OAuthProvider) (goth.Provider, error) {
		return stripe.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/stripe/callback", publicURL)), nil
	})
	RegisterProvider("tiktok", func(p OAuthProvider) (goth.Provider, error) {
		return tiktok.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/tiktok/callback", publicURL)), nil
	})
	RegisterProvider("tumblr", func(p OAuthProvider) (goth.Provider, error) {
		useAuthorizeString := p.AdditionalDetails["use_authorize"]
		if useAuthorizeString != "" {
			useAuthorize, err := strconv.ParseBool(useAuthorizeString)
			if err != nil {
				return nil, errors.New("invalid use_authorize value for tumblr provider")
			}
			if useAuthorize {
				return tumblr.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/tumblr/callback", publicURL)), nil
			}
		}
		return tumblr.NewAuthenticate(p.Key, p.Secret, fmt.Sprintf("%s/auth/tumblr/callback", publicURL)), nil
	})
	RegisterProvider("twitch", func(p OAuthProvider) (goth.Provider, error) {
		return twitch.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/twitch/callback", publicURL)), nil
	})
	RegisterProvider("twitter", func(p OAuthProvider) (goth.Provider, error) {
		useAuthorizeString := p.AdditionalDetails["use_authorize"]
		if useAuthorizeString != "" {
			useAuthorize, err := strconv.ParseBool(useAuthorizeString)
			if err != nil {
				return nil, errors.New("invalid use_authorize value for twitter provider")
			}
			if useAuthorize {
				return twitterv2.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/twitter/callback", publicURL)), nil
			}
		}
		return twitterv2.NewAuthenticate(p.Key, p.Secret, fmt.Sprintf("%s/auth/twitter/callback", publicURL)), nil
	})
	RegisterProvider("uber", func(p OAuthProvider) (goth.Provider, error) {
		return uber.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/uber/callback", publicURL)), nil
	})
	RegisterProvider("vk", func(p OAuthProvider) (goth.Provider, error) {
		return vk.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/vk/callback", publicURL)), nil
	})
	RegisterProvider("wechat", func(p OAuthProvider) (goth.Provider, error) {
		lang := p.AdditionalDetails["lang"]
		if lang == "" {
			lang = string(wechat.WECHAT_LANG_CN)
		}
		return wechat.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/auth/wechat/callback", publicURL),
			wechat.WechatLangType(lang),
		), nil
	})
	RegisterProvider("wecom", func(p OAuthProvider) (goth.Provider, error) {
		agentID := p.AdditionalDetails["agent_id"]
		if agentID == "" {
			return nil, errors.New("missing agent_id for wecom provider")
		}
		return wecom.New(p.Key, p.Secret, agentID, fmt.Sprintf("%s/auth/wecom/callback", publicURL)), nil
	})
	RegisterProvider("wepay", func(p OAuthProvider) (goth.Provider, error) {
		return wepay.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/wepay/callback", publicURL), "view_user"), nil
	})
	RegisterProvider("xero", func(p OAuthProvider) (goth.Provider, error) {
		return xero.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/xero/callback", publicURL)), nil
	})
	RegisterProvider("yahoo", func(p OAuthProvider) (goth.Provider, error) {
		return yahoo.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/yahoo/callback", publicURL)), nil
	})
	RegisterProvider("yammer", func(p OAuthProvider) (goth.Provider, error) {
		return yammer.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/yammer/callback", publicURL)), nil
	})
	RegisterProvider("yandex", func(p OAuthProvider) (goth.Provider, error) {
		return yandex.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/yandex/callback", publicURL)), nil
	})
	RegisterProvider("zoom", func(p OAuthProvider) (goth.Provider, error) {
		return zoom.New(p.Key, p.Secret, fmt.Sprintf("%s/auth/zoom/callback", publicURL), "read:user"), nil
	})
}

var providerNames = map[string]string{
	"amazon":          "Amazon",
	"apple":           "Apple",
	"auth0":           "Auth0",
	"azuread":         "Azure AD",
	"battlenet":       "Battle.net",
	"bitbucket":       "Bitbucket",
	"bitly":           "Bitly",
	"box":             "Box",
	"classlink":       "ClassLink",
	"cloudfoundry":    "Cloud Foundry",
	"cognito":         "Cognito",
	"dailymotion":     "Dailymotion",
	"deezer":          "Deezer",
	"digitalocean":    "DigitalOcean",
	"discord":         "Discord",
	"dropbox":         "Dropbox",
	"eveonline":       "Eve Online",
	"facebook":        "Facebook",
	"fitbit":          "Fitbit",
	"gitea":           "Gitea",
	"github":          "GitHub",
	"gitlab":          "Gitlab",
	"google":          "Google",
	"heroku":          "Heroku",
	"hubspot":         "HubSpot",
	"influxcloud":     "InfluxCloud",
	"instagram":       "Instagram",
	"intercom":        "Intercom",
	"kakao":           "Kakao",
	"lark":            "Lark",
	"lastfm":          "Last.fm",
	"line":            "LINE",
	"linkedin":        "LinkedIn",
	"mastodon":        "Mastodon",
	"meetup":          "Meetup",
	"microsoftonline": "Microsoft",
	"naver":           "Naver",
	"nextcloud":       "Nextcloud",
	"okta":            "Okta",
	"onedrive":        "OneDrive",
	"openid-connect":  "OpenID Connect",
	"oura":            "Oura",
	"patreon":         "Patreon",
	"paypal":          "PayPal",
	"salesforce":      "Salesforce",
	"seatalk":         "SeaTalk",
	"shopify":         "Shopify",
	"slack":           "Slack",
	"soundcloud":      "SoundCloud",
	"spotify":         "Spotify",
	"steam":           "Steam",
	"strava":          "Strava",
	"stripe":          "Stripe",
	"tiktok":          "TikTok",
	"tumblr":          "Tumblr",
	"twitch":          "Twitch",
	"twitter":         "Twitter",
	"uber":            "Uber",
	"vk":              "VK",
	"wechat":          "WeChat",
	"wecom":           "WeCom",
	"wepay":           "WePay",
	"xero":            "Xero",
	"yahoo":           "Yahoo",
	"yammer":          "Yammer",
	"yandex":          "Yandex",
	"zoom":            "Zoom",
}
