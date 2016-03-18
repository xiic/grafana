package api

import (
	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/middleware"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/setting"
)

func setIndexViewData(c *middleware.Context) (*dtos.IndexViewData, error) {
	settings, err := getFrontendSettingsMap(c)
	if err != nil {
		return nil, err
	}

	var data = dtos.IndexViewData{
		User: &dtos.CurrentUser{
			Id:             c.UserId,
			IsSignedIn:     c.IsSignedIn,
			Login:          c.Login,
			Email:          c.Email,
			Name:           c.Name,
			LightTheme:     c.Theme == "light",
			OrgId:          c.OrgId,
			OrgName:        c.OrgName,
			OrgRole:        c.OrgRole,
			GravatarUrl:    dtos.GetGravatarUrl(c.Email),
			IsGrafanaAdmin: c.IsGrafanaAdmin,
		},
		Settings:           settings,
		AppUrl:             setting.AppUrl,
		AppSubUrl:          setting.AppSubUrl,
		GoogleAnalyticsId:  setting.GoogleAnalyticsId,
		GoogleTagManagerId: setting.GoogleTagManagerId,
	}

	if setting.DisableGravatar {
		data.User.GravatarUrl = setting.AppSubUrl + "/public/img/transparent.png"
	}

	if len(data.User.Name) == 0 {
		data.User.Name = data.User.Login
	}

	themeUrlParam := c.Query("theme")
	if themeUrlParam == "light" {
		data.User.LightTheme = true
	}

	dashboardChildNavs := []*dtos.NavLink{
		{Text: "Home", Url: setting.AppSubUrl + "/"},
		{Text: "Playlists", Url: setting.AppSubUrl + "/playlists"},
		{Text: "Snapshots", Url: setting.AppSubUrl + "/dashboard/snapshots"},
	}

	if c.OrgRole == m.ROLE_ADMIN || c.OrgRole == m.ROLE_EDITOR {
		dashboardChildNavs = append(dashboardChildNavs, &dtos.NavLink{Divider: true})
		dashboardChildNavs = append(dashboardChildNavs, &dtos.NavLink{Text: "New", Url: setting.AppSubUrl + "/dashboard/new"})
		dashboardChildNavs = append(dashboardChildNavs, &dtos.NavLink{Text: "Import", Url: setting.AppSubUrl + "/import/dashboard"})
	}

	data.MainNavLinks = append(data.MainNavLinks, &dtos.NavLink{
		Text:     "Dashboards",
		Icon:     "icon-gf icon-gf-dashboard",
		Url:      setting.AppSubUrl + "/",
		Children: dashboardChildNavs,
	})

	if c.OrgRole == m.ROLE_ADMIN {
		data.MainNavLinks = append(data.MainNavLinks, &dtos.NavLink{
			Text: "Data Sources",
			Icon: "icon-gf icon-gf-datasources",
			Url:  setting.AppSubUrl + "/datasources",
		})

		data.MainNavLinks = append(data.MainNavLinks, &dtos.NavLink{
			Text: "Plugins",
			Icon: "icon-gf icon-gf-apps",
			Url:  setting.AppSubUrl + "/plugins",
		})
	}

	enabledPlugins, err := plugins.GetEnabledPlugins(c.OrgId)
	if err != nil {
		return nil, err
	}

	for _, plugin := range enabledPlugins.Apps {
		if plugin.Pinned {
			pageLink := &dtos.NavLink{
				Text: plugin.Name,
				Url:  setting.AppSubUrl + "/plugins/" + plugin.Id + "/edit",
				Img:  plugin.Info.Logos.Small,
			}

			for _, page := range plugin.Pages {
				if !page.SuppressNav {
					pageLink.Children = append(pageLink.Children, &dtos.NavLink{
						Url:  setting.AppSubUrl + "/plugins/" + plugin.Id + "/page/" + page.Slug,
						Text: page.Name,
					})
				}
			}

			data.MainNavLinks = append(data.MainNavLinks, pageLink)
		}
	}

	if c.IsGrafanaAdmin {
		data.MainNavLinks = append(data.MainNavLinks, &dtos.NavLink{
			Text: "Admin",
			Icon: "fa fa-fw fa-cogs",
			Url:  setting.AppSubUrl + "/admin",
			Children: []*dtos.NavLink{
				{Text: "Global Users", Icon: "fa fa-fw fa-cogs", Url: setting.AppSubUrl + "/admin/users"},
				{Text: "Global Orgs", Icon: "fa fa-fw fa-cogs", Url: setting.AppSubUrl + "/admin/orgs"},
				{Text: "Server Settings", Icon: "fa fa-fw fa-cogs", Url: setting.AppSubUrl + "/admin/settings"},
				{Text: "Server Stats", Icon: "fa-fw fa-cogs", Url: setting.AppSubUrl + "/admin/stats"},
			},
		})
	}

	return &data, nil
}

func Index(c *middleware.Context) {
	if data, err := setIndexViewData(c); err != nil {
		c.Handle(500, "Failed to get settings", err)
		return
	} else {
		c.HTML(200, "index", data)
	}
}

func NotFoundHandler(c *middleware.Context) {
	if c.IsApiRequest() {
		c.JsonApiErr(404, "Not found", nil)
		return
	}

	if data, err := setIndexViewData(c); err != nil {
		c.Handle(500, "Failed to get settings", err)
		return
	} else {
		c.HTML(404, "index", data)
	}
}
