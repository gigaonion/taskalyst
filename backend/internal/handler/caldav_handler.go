package handler

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gigaonion/taskalyst/backend/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type CalDavHandler struct {
	u               usecase.CalDavUsecase
	calendarUsecase usecase.CalendarUsecase
}

func NewCalDavHandler(u usecase.CalDavUsecase, calendarUsecase usecase.CalendarUsecase) *CalDavHandler {
	return &CalDavHandler{u: u, calendarUsecase: calendarUsecase}
}

// --- WebDAV / CalDAV XML Structures ---

type Multistatus struct {
	XMLName   xml.Name   `xml:"DAV: multistatus"`
	Responses []Response `xml:"response"`
}

type Response struct {
	Href      string     `xml:"DAV: href"`
	Propstats []Propstat `xml:"DAV: propstat"`
}

type Propstat struct {
	Prop   Prop   `xml:"DAV: prop"`
	Status string `xml:"DAV: status"`
}

type Prop struct {
	Displayname           string                 `xml:"DAV: displayname,omitempty"`
	Resourcetype          *Resourcetype          `xml:"DAV: resourcetype,omitempty"`
	CurrentUserPrincipal  *HrefProp              `xml:"DAV: current-user-principal,omitempty"`
	CalendarHomeSet       *HrefProp              `xml:"urn:ietf:params:xml:ns:caldav calendar-home-set,omitempty"`
	SupportedCalendarComp *SupportedCalendarComp `xml:"urn:ietf:params:xml:ns:caldav supported-calendar-component-set,omitempty"`
	CalendarDescription   string                 `xml:"urn:ietf:params:xml:ns:caldav calendar-description,omitempty"`
	GetContentType        string                 `xml:"DAV: getcontenttype,omitempty"`
	GetETag               string                 `xml:"DAV: getetag,omitempty"`
	GetContentLength      int64                  `xml:"DAV: getcontentlength,omitempty"`
	CalendarData          string                 `xml:"urn:ietf:params:xml:ns:caldav calendar-data,omitempty"`
}

type Resourcetype struct {
	Collection *struct{} `xml:"DAV: collection,omitempty"`
	Principal  *struct{} `xml:"DAV: principal,omitempty"`
	Calendar   *struct{} `xml:"urn:ietf:params:xml:ns:caldav calendar,omitempty"`
}

type HrefProp struct {
	Href string `xml:"DAV: href"`
}

type SupportedCalendarComp struct {
	Components []Comp `xml:"comp"`
}

type Comp struct {
	Name string `xml:"name,attr"`
}

// Propfind request body
type PropfindRequest struct {
	XMLName xml.Name  `xml:"DAV: propfind"`
	Prop    *Prop     `xml:"prop,omitempty"`
	AllProp *struct{} `xml:"allprop,omitempty"`
}

// Report request
type Report struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:caldav calendar-query"`
	Prop    Prop     `xml:"DAV: prop"`
	Filter  Filter   `xml:"filter"`
}

type Filter struct {
	CompFilter CompFilter `xml:"comp-filter"`
}

type CompFilter struct {
	Name       string      `xml:"name,attr"`
	TimeRange  *TimeRange  `xml:"time-range,omitempty"`
	CompFilter *CompFilter `xml:"comp-filter,omitempty"`
}

type TimeRange struct {
	Start string `xml:"start,attr,omitempty"`
	End   string `xml:"end,attr,omitempty"`
}

// --- Handler Methods ---

func (h *CalDavHandler) Options(c echo.Context) error {
	h.setDavHeaders(c)
	return c.NoContent(http.StatusOK)
}

func (h *CalDavHandler) setDavHeaders(c echo.Context) {
	c.Response().Header().Set("Allow", "OPTIONS, GET, HEAD, DELETE, PROPFIND, PUT, PROPPATCH, REPORT, MKCOL, MKCALENDAR")
	c.Response().Header().Set("DAV", "1, 2, calendar-access, calendar-proxy")
}

// PrincipalDiscovery handles PROPFIND /dav/principals/
func (h *CalDavHandler) PrincipalDiscovery(c echo.Context) error {
	if c.Request().Method == "OPTIONS" {
		return h.Options(c)
	}

	userID := getUserID(c)
	if userID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	requestedProps := h.parsePropfindRequest(c)
	principalHref := fmt.Sprintf("/dav/principals/%s/", userID.String())

	res := Multistatus{
		Responses: []Response{
			{
				Href: "/dav/principals/",
				Propstats: h.buildPropstats(requestedProps, Prop{
					Resourcetype:         &Resourcetype{Collection: &struct{}{}},
					CurrentUserPrincipal: &HrefProp{Href: principalHref},
				}),
			},
			{
				Href: principalHref,
				Propstats: h.buildPropstats(requestedProps, Prop{
					Resourcetype:         &Resourcetype{Principal: &struct{}{}, Collection: &struct{}{}},
					CurrentUserPrincipal: &HrefProp{Href: principalHref},
					Displayname:          userID.String(),
				}),
			},
		},
	}

	return h.xmlResponse(c, http.StatusMultiStatus, res)
}

// Principal handles PROPFIND /dav/principals/:userID/
func (h *CalDavHandler) Principal(c echo.Context) error {
	if c.Request().Method == "OPTIONS" {
		return h.Options(c)
	}

	userIDStr := c.Param("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	currentUserID := getUserID(c)
	if userID != currentUserID {
		return echo.NewHTTPError(http.StatusForbidden)
	}

	requestedProps := h.parsePropfindRequest(c)
	principalHref := fmt.Sprintf("/dav/principals/%s/", userID.String())
	calendarHomeHref := fmt.Sprintf("/dav/calendars/%s/", userID.String())

	res := Multistatus{
		Responses: []Response{
			{
				Href: principalHref,
				Propstats: h.buildPropstats(requestedProps, Prop{
					Resourcetype:         &Resourcetype{Principal: &struct{}{}, Collection: &struct{}{}},
					Displayname:          userIDStr,
					CurrentUserPrincipal: &HrefProp{Href: principalHref},
					CalendarHomeSet:      &HrefProp{Href: calendarHomeHref},
				}),
			},
		},
	}

	return h.xmlResponse(c, http.StatusMultiStatus, res)
}

// CalendarHome handles PROPFIND /dav/calendars/:userID/
func (h *CalDavHandler) CalendarHome(c echo.Context) error {
	if c.Request().Method == "OPTIONS" {
		return h.Options(c)
	}

	userID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	currentUserID := getUserID(c)
	if userID != currentUserID {
		return echo.NewHTTPError(http.StatusForbidden)
	}

	calendars, err := h.u.GetCalendars(c.Request().Context(), userID)
	if err != nil {
		return HandleError(c, err)
	}

	requestedProps := h.parsePropfindRequest(c)
	homeHref := fmt.Sprintf("/dav/calendars/%s/", userID.String())

	responses := []Response{
		{
			Href: homeHref,
			Propstats: h.buildPropstats(requestedProps, Prop{
				Resourcetype: &Resourcetype{Collection: &struct{}{}},
			}),
		},
	}

	for _, cal := range calendars {
		calHref := fmt.Sprintf("/dav/calendars/%s/%s/", userID.String(), cal.ID.String())
		responses = append(responses, Response{
			Href: calHref,
			Propstats: h.buildPropstats(requestedProps, Prop{
				Displayname:  cal.Name,
				Resourcetype: &Resourcetype{Collection: &struct{}{}, Calendar: &struct{}{}},
				SupportedCalendarComp: &SupportedCalendarComp{
					Components: []Comp{{Name: "VEVENT"}, {Name: "VTODO"}},
				},
			}),
		})
	}

	return h.xmlResponse(c, http.StatusMultiStatus, Multistatus{Responses: responses})
}

// CalendarCollection handles PROPFIND, REPORT /dav/calendars/:userID/:calendarID
func (h *CalDavHandler) CalendarCollection(c echo.Context) error {
	if c.Request().Method == "OPTIONS" {
		return h.Options(c)
	}

	userID, _ := uuid.Parse(c.Param("userID"))
	calendarID, err := uuid.Parse(c.Param("calendarID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid calendar id")
	}

	if c.Request().Method == "REPORT" {
		return h.HandleReport(c, userID, calendarID)
	}

	// PROPFIND
	requestedProps := h.parsePropfindRequest(c)
	depth := c.Request().Header.Get("Depth")

	cal, err := h.u.GetCalendar(c.Request().Context(), userID, calendarID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	calHref := fmt.Sprintf("/dav/calendars/%s/%s/", userID.String(), calendarID.String())

	responses := []Response{
		{
			Href: calHref,
			Propstats: h.buildPropstats(requestedProps, Prop{
				Displayname:  cal.Name,
				Resourcetype: &Resourcetype{Collection: &struct{}{}, Calendar: &struct{}{}},
				SupportedCalendarComp: &SupportedCalendarComp{
					Components: []Comp{{Name: "VEVENT"}, {Name: "VTODO"}},
				},
			}),
		},
	}

	// RFC 4918: If Depth is omitted, it should be treated as infinity.
	if depth == "1" || depth == "infinity" || depth == "" {
		// List all events and tasks
		events, _ := h.u.GetEventsByRange(c.Request().Context(), userID, calendarID, time.Time{}, time.Time{})
		tasks, _ := h.u.GetTasksByRange(c.Request().Context(), userID, calendarID, time.Time{}, time.Time{})

		for _, e := range events {
			href := fmt.Sprintf("%s%s.ics", calHref, e.IcalUid.String)
			responses = append(responses, Response{
				Href: href,
				Propstats: h.buildPropstats(requestedProps, Prop{
					GetContentType: "text/calendar; charset=utf-8",
					GetETag:        fmt.Sprintf("\"%s\"", e.Etag.String),
				}),
			})
		}
		for _, t := range tasks {
			href := fmt.Sprintf("%s%s.ics", calHref, t.IcalUid.String)
			responses = append(responses, Response{
				Href: href,
				Propstats: h.buildPropstats(requestedProps, Prop{
					GetContentType: "text/calendar; charset=utf-8",
					GetETag:        fmt.Sprintf("\"%s\"", t.Etag.String),
				}),
			})
		}
	}

	return h.xmlResponse(c, http.StatusMultiStatus, Multistatus{Responses: responses})
}

// CalendarResource handles GET, PUT, DELETE /dav/calendars/:userID/:calendarID/:resource
func (h *CalDavHandler) CalendarResource(c echo.Context) error {
	if c.Request().Method == "OPTIONS" {
		return h.Options(c)
	}

	userID, _ := uuid.Parse(c.Param("userID"))
	calendarID, _ := uuid.Parse(c.Param("calendarID"))
	resourceName := c.Param("resource")

	if !strings.HasSuffix(resourceName, ".ics") {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	icalUID := strings.TrimSuffix(resourceName, ".ics")

	switch c.Request().Method {
	case "GET":
		// Try event first
		ical, err := h.u.ExportEventToICal(c.Request().Context(), userID, icalUID)
		if err != nil {
			// Try task
			ical, err = h.u.ExportTaskToICal(c.Request().Context(), userID, icalUID)
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound)
			}
		}
		c.Response().Header().Set("Content-Type", "text/calendar; charset=utf-8")
		return c.String(http.StatusOK, ical)

	case "PUT":
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "failed to read body")
		}
		if err := h.u.ImportFromICal(c.Request().Context(), userID, calendarID, string(body)); err != nil {
			return HandleError(c, err)
		}
		return c.NoContent(http.StatusCreated)

	case "DELETE":
		if err := h.u.DeleteResource(c.Request().Context(), userID, icalUID); err != nil {
			return HandleError(c, err)
		}
		return c.NoContent(http.StatusNoContent)

	case "PROPFIND":
		requestedProps := h.parsePropfindRequest(c)
		return h.xmlResponse(c, http.StatusMultiStatus, Multistatus{
			Responses: []Response{
				{
					Href: c.Request().URL.Path,
					Propstats: h.buildPropstats(requestedProps, Prop{
						GetContentType: "text/calendar; charset=utf-8",
					}),
				},
			},
		})
	}

	return echo.NewHTTPError(http.StatusMethodNotAllowed)
}

func (h *CalDavHandler) HandleReport(c echo.Context, userID, calendarID uuid.UUID) error {
	var report Report
	if err := xml.NewDecoder(c.Request().Body).Decode(&report); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported report")
	}

	// Extract time range if present
	var start, end time.Time
	if report.Filter.CompFilter.CompFilter != nil && report.Filter.CompFilter.CompFilter.TimeRange != nil {
		tr := report.Filter.CompFilter.CompFilter.TimeRange
		if tr.Start != "" {
			start, _ = time.Parse("20060102T150405Z", tr.Start)
		}
		if tr.End != "" {
			end, _ = time.Parse("20060102T150405Z", tr.End)
		}
	}

	events, _ := h.u.GetEventsByRange(c.Request().Context(), userID, calendarID, start, end)
	tasks, _ := h.u.GetTasksByRange(c.Request().Context(), userID, calendarID, start, end)

	responses := []Response{}
	calHref := fmt.Sprintf("/dav/calendars/%s/%s/", userID.String(), calendarID.String())

	for _, e := range events {
		ical, _ := h.u.ExportEventToICal(c.Request().Context(), userID, e.IcalUid.String)
		responses = append(responses, Response{
			Href: fmt.Sprintf("%s%s.ics", calHref, e.IcalUid.String),
			Propstats: []Propstat{
				{
					Prop: Prop{
						GetContentType: "text/calendar; charset=utf-8",
						GetETag:        fmt.Sprintf("\"%s\"", e.Etag.String),
						CalendarData:   ical,
					},
					Status: "HTTP/1.1 200 OK",
				},
			},
		})
	}
	for _, t := range tasks {
		ical, _ := h.u.ExportTaskToICal(c.Request().Context(), userID, t.IcalUid.String)
		responses = append(responses, Response{
			Href: fmt.Sprintf("%s%s.ics", calHref, t.IcalUid.String),
			Propstats: []Propstat{
				{
					Prop: Prop{
						GetContentType: "text/calendar; charset=utf-8",
						GetETag:        fmt.Sprintf("\"%s\"", t.Etag.String),
						CalendarData:   ical,
					},
					Status: "HTTP/1.1 200 OK",
				},
			},
		})
	}

	return h.xmlResponse(c, http.StatusMultiStatus, Multistatus{Responses: responses})
}

// --- Helpers ---

func (h *CalDavHandler) parsePropfindRequest(c echo.Context) *Prop {
	if c.Request().ContentLength <= 0 {
		return nil // Return all properties if body is empty (RFC 4918 prefers allprop)
	}

	var req PropfindRequest
	if err := xml.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return nil
	}
	if req.AllProp != nil {
		return nil
	}
	return req.Prop
}

func (h *CalDavHandler) buildPropstats(requested *Prop, available Prop) []Propstat {
	if requested == nil {
		// Return all available properties
		return []Propstat{{Prop: available, Status: "HTTP/1.1 200 OK"}}
	}

	found := Prop{}
	notFound := Prop{}
	hasFound := false
	hasNotFound := false

	// Check each requested property
	// Displayname
	if requested.Displayname != "" {
		if available.Displayname != "" {
			found.Displayname = available.Displayname
			hasFound = true
		} else {
			notFound.Displayname = " " // Empty tag to indicate missing
			hasNotFound = true
		}
	}

	// Resourcetype
	if requested.Resourcetype != nil {
		if available.Resourcetype != nil {
			found.Resourcetype = available.Resourcetype
			hasFound = true
		} else {
			notFound.Resourcetype = &Resourcetype{}
			hasNotFound = true
		}
	}

	// CurrentUserPrincipal
	if requested.CurrentUserPrincipal != nil {
		if available.CurrentUserPrincipal != nil {
			found.CurrentUserPrincipal = available.CurrentUserPrincipal
			hasFound = true
		} else {
			notFound.CurrentUserPrincipal = &HrefProp{}
			hasNotFound = true
		}
	}

	// CalendarHomeSet
	if requested.CalendarHomeSet != nil {
		if available.CalendarHomeSet != nil {
			found.CalendarHomeSet = available.CalendarHomeSet
			hasFound = true
		} else {
			notFound.CalendarHomeSet = &HrefProp{}
			hasNotFound = true
		}
	}

	// SupportedCalendarComp
	if requested.SupportedCalendarComp != nil {
		if available.SupportedCalendarComp != nil {
			found.SupportedCalendarComp = available.SupportedCalendarComp
			hasFound = true
		} else {
			notFound.SupportedCalendarComp = &SupportedCalendarComp{}
			hasNotFound = true
		}
	}

	// GetContentType
	if requested.GetContentType != "" {
		if available.GetContentType != "" {
			found.GetContentType = available.GetContentType
			hasFound = true
		} else {
			notFound.GetContentType = " "
			hasNotFound = true
		}
	}

	// GetETag
	if requested.GetETag != "" {
		if available.GetETag != "" {
			found.GetETag = available.GetETag
			hasFound = true
		} else {
			notFound.GetETag = " "
			hasNotFound = true
		}
	}

	// CalendarData
	if requested.CalendarData != "" {
		if available.CalendarData != "" {
			found.CalendarData = available.CalendarData
			hasFound = true
		} else {
			notFound.CalendarData = " "
			hasNotFound = true
		}
	}

	// Always include resourcetype if it's there as it's fundamental and found
	if found.Resourcetype == nil && available.Resourcetype != nil && hasFound {
		found.Resourcetype = available.Resourcetype
	}

	var propstats []Propstat
	if hasFound {
		propstats = append(propstats, Propstat{Prop: found, Status: "HTTP/1.1 200 OK"})
	}
	if hasNotFound {
		propstats = append(propstats, Propstat{Prop: notFound, Status: "HTTP/1.1 404 Not Found"})
	}

	return propstats
}

func (h *CalDavHandler) xmlResponse(c echo.Context, code int, data interface{}) error {
	c.Response().Header().Set("Content-Type", "application/xml; charset=utf-8")
	c.Response().WriteHeader(code)
	if _, err := c.Response().Write([]byte(xml.Header)); err != nil {
		return err
	}
	return xml.NewEncoder(c.Response()).Encode(data)
}

// Legacy methods
func (h *CalDavHandler) GetCalendar(c echo.Context) error {
	userID, _ := uuid.Parse(c.Param("userID"))
	calendarID, err := uuid.Parse(c.Param("calendarID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid calendar id")
	}

	ical, err := h.u.ExportCalendarToICal(c.Request().Context(), userID, calendarID)
	if err != nil {
		return HandleError(c, err)
	}

	c.Response().Header().Set("Content-Type", "text/calendar; charset=utf-8")
	return c.String(http.StatusOK, ical)
}
