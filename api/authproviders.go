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

type AuthProvider struct {
	ID                string            `yaml:"id"`
	Key               string            `yaml:"key,alias=client_id"`
	Secret            string            `yaml:"secret,alias=client_secret"`
	AdditionalDetails map[string]string `yaml:",inline"`
}

func (p AuthProvider) Name() string {
	return providerNames[p.ID]
}

func (p AuthProvider) GothProvider() (goth.Provider, error) {
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

type providerFactory func(AuthProvider) (goth.Provider, error)

var providerFactories = map[string]providerFactory{}

func RegisterProvider(id string, factory providerFactory) {
	providerFactories[id] = factory
}

//nolint:gocyclo,funlen // This function is long but we need a way to register all the providers
func RegisterBuiltInProviders(publicURL string) {
	RegisterProvider("amazon", func(p AuthProvider) (goth.Provider, error) {
		return amazon.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/amazon", publicURL)), nil
	})
	RegisterProvider("apple", func(p AuthProvider) (goth.Provider, error) {
		return apple.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/api/auth/callback/apple", publicURL),
			nil,
			apple.ScopeName,
			apple.ScopeEmail,
		), nil
	})
	RegisterProvider("auth0", func(p AuthProvider) (goth.Provider, error) {
		domain := p.AdditionalDetails["domain"]
		if domain == "" {
			return nil, errors.New("missing domain for auth0 provider")
		}
		return auth0.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/auth0", publicURL), domain), nil
	})
	RegisterProvider("azuread", func(p AuthProvider) (goth.Provider, error) {
		return azureadv2.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/api/auth/callback/azuread", publicURL),
			azureadv2.ProviderOptions{},
		), nil
	})
	RegisterProvider("battlenet", func(p AuthProvider) (goth.Provider, error) {
		return battlenet.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/battlenet", publicURL)), nil
	})
	RegisterProvider("bitbucket", func(p AuthProvider) (goth.Provider, error) {
		return bitbucket.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/bitbucket", publicURL)), nil
	})
	RegisterProvider("bitly", func(p AuthProvider) (goth.Provider, error) {
		return bitly.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/bitly", publicURL)), nil
	})
	RegisterProvider("box", func(p AuthProvider) (goth.Provider, error) {
		return box.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/box", publicURL)), nil
	})
	RegisterProvider("classlink", func(p AuthProvider) (goth.Provider, error) {
		return classlink.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/classlink", publicURL)), nil
	})
	RegisterProvider("cloudfoundry", func(p AuthProvider) (goth.Provider, error) {
		authURL := p.AdditionalDetails["auth_url"]
		if authURL == "" {
			return nil, errors.New("missing auth_url for cloudfoundry provider")
		}
		return cloudfoundry.New(
			authURL,
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/api/auth/callback/cloudfoundry", publicURL),
		), nil
	})
	RegisterProvider("cognito", func(p AuthProvider) (goth.Provider, error) {
		baseURL := p.AdditionalDetails["base_url"]
		if baseURL != "" {
			return cognito.New(p.Key, p.Secret, baseURL, fmt.Sprintf("%s/api/auth/callback/cognito", publicURL)), nil
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
			fmt.Sprintf("%s/api/auth/callback/cognito", publicURL),
			authURL,
			tokenURL,
			issuerURL,
			profileURL,
		), nil
	})
	RegisterProvider("dailymotion", func(p AuthProvider) (goth.Provider, error) {
		return dailymotion.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/api/auth/callback/dailymotion", publicURL),
			"email",
		), nil
	})
	RegisterProvider("deezer", func(p AuthProvider) (goth.Provider, error) {
		return deezer.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/deezer", publicURL), "email"), nil
	})
	RegisterProvider("digitalocean", func(p AuthProvider) (goth.Provider, error) {
		return digitalocean.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/api/auth/callback/digitalocean", publicURL),
			"read",
		), nil
	})
	RegisterProvider("discord", func(p AuthProvider) (goth.Provider, error) {
		return discord.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/api/auth/callback/discord", publicURL),
			discord.ScopeIdentify,
			discord.ScopeEmail,
		), nil
	})
	RegisterProvider("dropbox", func(p AuthProvider) (goth.Provider, error) {
		return dropbox.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/dropbox", publicURL)), nil
	})
	RegisterProvider("eveonline", func(p AuthProvider) (goth.Provider, error) {
		return eveonline.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/eveonline", publicURL)), nil
	})
	RegisterProvider("facebook", func(p AuthProvider) (goth.Provider, error) {
		return facebook.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/facebook", publicURL)), nil
	})
	RegisterProvider("fitbit", func(p AuthProvider) (goth.Provider, error) {
		return fitbit.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/fitbit", publicURL)), nil
	})
	RegisterProvider("gitea", func(p AuthProvider) (goth.Provider, error) {
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
				fmt.Sprintf("%s/api/auth/callback/gitea", publicURL),
				authURL,
				tokenURL,
				profileURL,
			), nil
		}
		return gitea.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/gitea", publicURL)), nil
	})
	RegisterProvider("github", func(p AuthProvider) (goth.Provider, error) {
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
				fmt.Sprintf("%s/api/auth/callback/github", publicURL),
				authURL,
				tokenURL,
				profileURL,
				emailURL,
			), nil
		}
		return github.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/github", publicURL)), nil
	})
	RegisterProvider("gitlab", func(p AuthProvider) (goth.Provider, error) {
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
				fmt.Sprintf("%s/api/auth/callback/gitlab", publicURL),
				authURL,
				tokenURL,
				profileURL,
			), nil
		}
		return gitlab.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/gitlab", publicURL)), nil
	})
	RegisterProvider("google", func(p AuthProvider) (goth.Provider, error) {
		return google.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/google", publicURL)), nil
	})
	RegisterProvider("heroku", func(p AuthProvider) (goth.Provider, error) {
		return heroku.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/heroku", publicURL)), nil
	})
	RegisterProvider("hubspot", func(p AuthProvider) (goth.Provider, error) {
		return hubspot.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/hubspot", publicURL)), nil
	})
	RegisterProvider("influxcloud", func(p AuthProvider) (goth.Provider, error) {
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
				fmt.Sprintf("%s/api/auth/callback/influxcloud", publicURL),
				authURL,
				tokenURL,
				userAPIEndpoint,
				"userscope",
			), nil
		}
		return influxcloud.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/api/auth/callback/influxcloud", publicURL),
			"userscope",
		), nil
	})
	RegisterProvider("instagram", func(p AuthProvider) (goth.Provider, error) {
		return instagram.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/instagram", publicURL)), nil
	})
	RegisterProvider("intercom", func(p AuthProvider) (goth.Provider, error) {
		return intercom.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/intercom", publicURL)), nil
	})
	RegisterProvider("kakao", func(p AuthProvider) (goth.Provider, error) {
		return kakao.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/kakao", publicURL)), nil
	})
	RegisterProvider("lark", func(p AuthProvider) (goth.Provider, error) {
		return lark.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/lark", publicURL)), nil
	})
	RegisterProvider("lastfm", func(p AuthProvider) (goth.Provider, error) {
		return lastfm.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/lastfm", publicURL)), nil
	})
	RegisterProvider("line", func(p AuthProvider) (goth.Provider, error) {
		return line.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/api/auth/callback/line", publicURL),
			"profile",
			"openid",
			"email",
		), nil
	})
	RegisterProvider("linkedin", func(p AuthProvider) (goth.Provider, error) {
		return linkedin.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/linkedin", publicURL)), nil
	})
	RegisterProvider("mastodon", func(p AuthProvider) (goth.Provider, error) {
		instanceURL := p.AdditionalDetails["instance_url"]
		if instanceURL != "" {
			return mastodon.NewCustomisedURL(
				p.Key,
				p.Secret,
				fmt.Sprintf("%s/api/auth/callback/mastodon", publicURL),
				instanceURL,
				"read:accounts",
			), nil
		}
		return mastodon.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/api/auth/callback/mastodon", publicURL),
			"read:accounts",
		), nil
	})
	RegisterProvider("meetup", func(p AuthProvider) (goth.Provider, error) {
		return meetup.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/meetup", publicURL)), nil
	})
	RegisterProvider("microsoftonline", func(p AuthProvider) (goth.Provider, error) {
		return microsoftonline.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/microsoftonline", publicURL)), nil
	})
	RegisterProvider("naver", func(p AuthProvider) (goth.Provider, error) {
		return naver.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/naver", publicURL)), nil
	})
	RegisterProvider("nextcloud", func(p AuthProvider) (goth.Provider, error) {
		nextcloudURL := p.AdditionalDetails["nextcloud_url"]
		if nextcloudURL != "" {
			return nextcloud.NewCustomisedDNS(
				p.Key,
				p.Secret,
				fmt.Sprintf("%s/api/auth/callback/nextcloud", publicURL),
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
			fmt.Sprintf("%s/api/auth/callback/nextcloud", publicURL),
			authURL,
			tokenURL,
			profileURL,
		), nil
	})
	RegisterProvider("okta", func(p AuthProvider) (goth.Provider, error) {
		orgURL := p.AdditionalDetails["org_url"]
		if orgURL != "" {
			return okta.New(p.Key, p.Secret, orgURL, fmt.Sprintf("%s/api/auth/callback/okta", publicURL)), nil
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
			fmt.Sprintf("%s/api/auth/callback/okta", publicURL),
			authURL,
			tokenURL,
			issuerURL,
			profileURL,
		), nil
	})
	RegisterProvider("onedrive", func(p AuthProvider) (goth.Provider, error) {
		return onedrive.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/onedrive", publicURL)), nil
	})
	RegisterProvider("openid-connect", func(p AuthProvider) (goth.Provider, error) {
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
				fmt.Sprintf("%s/api/auth/callback/%s", publicURL, realName),
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
			fmt.Sprintf("%s/api/auth/callback/%s", publicURL, realName),
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
	RegisterProvider("oura", func(p AuthProvider) (goth.Provider, error) {
		return oura.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/api/auth/callback/oura", publicURL),
			oura.ScopeEmail,
			oura.ScopePersonal,
		), nil
	})
	RegisterProvider("patreon", func(p AuthProvider) (goth.Provider, error) {
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
				fmt.Sprintf("%s/api/auth/callback/patreon", publicURL),
				authURL,
				tokenURL,
				profileURL,
			), nil
		}
		return patreon.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/patreon", publicURL)), nil
	})
	RegisterProvider("paypal", func(p AuthProvider) (goth.Provider, error) {
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
				fmt.Sprintf("%s/api/auth/callback/paypal", publicURL),
				authURL,
				tokenURL,
				profileURL,
			), nil
		}
		return paypal.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/paypal", publicURL)), nil
	})
	RegisterProvider("salesforce", func(p AuthProvider) (goth.Provider, error) {
		return salesforce.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/salesforce", publicURL)), nil
	})
	RegisterProvider("seatalk", func(p AuthProvider) (goth.Provider, error) {
		return seatalk.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/seatalk", publicURL)), nil
	})
	RegisterProvider("shopify", func(p AuthProvider) (goth.Provider, error) {
		return shopify.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/api/auth/callback/shopify", publicURL),
			shopify.ScopeReadCustomers,
		), nil
	})
	RegisterProvider("slack", func(p AuthProvider) (goth.Provider, error) {
		return slack.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/slack", publicURL)), nil
	})
	RegisterProvider("soundcloud", func(p AuthProvider) (goth.Provider, error) {
		return soundcloud.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/soundcloud", publicURL)), nil
	})
	RegisterProvider("spotify", func(p AuthProvider) (goth.Provider, error) {
		return spotify.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/spotify", publicURL)), nil
	})
	RegisterProvider("steam", func(p AuthProvider) (goth.Provider, error) {
		return steam.New(p.Key, fmt.Sprintf("%s/api/auth/callback/steam", publicURL)), nil
	})
	RegisterProvider("strava", func(p AuthProvider) (goth.Provider, error) {
		return strava.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/strava", publicURL)), nil
	})
	RegisterProvider("stripe", func(p AuthProvider) (goth.Provider, error) {
		return stripe.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/stripe", publicURL)), nil
	})
	RegisterProvider("tiktok", func(p AuthProvider) (goth.Provider, error) {
		return tiktok.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/tiktok", publicURL)), nil
	})
	RegisterProvider("tumblr", func(p AuthProvider) (goth.Provider, error) {
		useAuthorizeString := p.AdditionalDetails["use_authorize"]
		if useAuthorizeString != "" {
			useAuthorize, err := strconv.ParseBool(useAuthorizeString)
			if err != nil {
				return nil, errors.New("invalid use_authorize value for tumblr provider")
			}
			if useAuthorize {
				return tumblr.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/tumblr", publicURL)), nil
			}
		}
		return tumblr.NewAuthenticate(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/tumblr", publicURL)), nil
	})
	RegisterProvider("twitch", func(p AuthProvider) (goth.Provider, error) {
		return twitch.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/twitch", publicURL)), nil
	})
	RegisterProvider("twitter", func(p AuthProvider) (goth.Provider, error) {
		useAuthorizeString := p.AdditionalDetails["use_authorize"]
		if useAuthorizeString != "" {
			useAuthorize, err := strconv.ParseBool(useAuthorizeString)
			if err != nil {
				return nil, errors.New("invalid use_authorize value for twitter provider")
			}
			if useAuthorize {
				return twitterv2.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/twitter", publicURL)), nil
			}
		}
		return twitterv2.NewAuthenticate(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/twitter", publicURL)), nil
	})
	RegisterProvider("uber", func(p AuthProvider) (goth.Provider, error) {
		return uber.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/uber", publicURL)), nil
	})
	RegisterProvider("vk", func(p AuthProvider) (goth.Provider, error) {
		return vk.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/vk", publicURL)), nil
	})
	RegisterProvider("wechat", func(p AuthProvider) (goth.Provider, error) {
		lang := p.AdditionalDetails["lang"]
		if lang == "" {
			lang = string(wechat.WECHAT_LANG_CN)
		}
		return wechat.New(
			p.Key,
			p.Secret,
			fmt.Sprintf("%s/api/auth/callback/wechat", publicURL),
			wechat.WechatLangType(lang),
		), nil
	})
	RegisterProvider("wecom", func(p AuthProvider) (goth.Provider, error) {
		agentID := p.AdditionalDetails["agent_id"]
		if agentID == "" {
			return nil, errors.New("missing agent_id for wecom provider")
		}
		return wecom.New(p.Key, p.Secret, agentID, fmt.Sprintf("%s/api/auth/callback/wecom", publicURL)), nil
	})
	RegisterProvider("wepay", func(p AuthProvider) (goth.Provider, error) {
		return wepay.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/wepay", publicURL), "view_user"), nil
	})
	RegisterProvider("xero", func(p AuthProvider) (goth.Provider, error) {
		return xero.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/xero", publicURL)), nil
	})
	RegisterProvider("yahoo", func(p AuthProvider) (goth.Provider, error) {
		return yahoo.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/yahoo", publicURL)), nil
	})
	RegisterProvider("yammer", func(p AuthProvider) (goth.Provider, error) {
		return yammer.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/yammer", publicURL)), nil
	})
	RegisterProvider("yandex", func(p AuthProvider) (goth.Provider, error) {
		return yandex.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/yandex", publicURL)), nil
	})
	RegisterProvider("zoom", func(p AuthProvider) (goth.Provider, error) {
		return zoom.New(p.Key, p.Secret, fmt.Sprintf("%s/api/auth/callback/zoom", publicURL), "read:user"), nil
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
