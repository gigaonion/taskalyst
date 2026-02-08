package handler

import (
	"encoding/xml"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/gigaonion/taskalyst/backend/internal/usecase"
)

type CalDavHandler struct {
	u usecase.CalDavUsecase
	calendarUsecase usecase.CalendarUsecase
}

func NewCalDavHandler(u usecase.CalDavUsecase, calendarUsecase usecase.CalendarUsecase) *CalDavHandler {
	return &CalDavHandler{u: u, calendarUsecase: calendarUsecase}
}

// DAV XML structures
type Propfind struct {
	XMLName xml.Name `xml:"DAV: propfind"`
	Prop    Prop     `xml:"prop"`
}

type Prop struct {
	XMLName xml.Name `xml:"DAV: prop"`
	Resourcetype *Resourcetype `xml:"resourcetype,omitempty"`
	CurrentUserPrincipal *CurrentUserPrincipal `xml:"current-user-principal,omitempty"`
	CalendarHomeSet *CalendarHomeSet `xml:"urn:ietf:params:xml:ns:caldav calendar-home-set,omitempty"`
}

type Resourcetype struct {
	Collection *struct{} `xml:"collection,omitempty"`
	Calendar   *struct{} `xml:"urn:ietf:params:xml:ns:caldav calendar,omitempty"`
}

type CurrentUserPrincipal struct {
	Href string `xml:"href"`
}

type CalendarHomeSet struct {
	Href string `xml:"href"`
}

func (h *CalDavHandler) Options(c echo.Context) error {
	c.Response().Header().Set("Allow", "OPTIONS, GET, HEAD, DELETE, PROPFIND, PUT, PROPPATCH, REPORT, MKCOL, MKCALENDAR")
	c.Response().Header().Set("DAV", "1, 2, calendar-access")
	return c.NoContent(http.StatusOK)
}

func (h *CalDavHandler) Propfind(c echo.Context) error {
	//userID := c.Get("user_id").(uuid.UUID)
	// Simplified response for now
	c.Response().Header().Set("Content-Type", "application/xml; charset=utf-8")
	return c.String(http.StatusMultiStatus, `<?xml version="1.0" encoding="utf-8" ?>
<d:multistatus xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
    <d:response>
        <d:href>`+c.Request().URL.Path+`</d:href>
        <d:propstat>
            <d:prop>
                <d:resourcetype><d:collection/></d:resourcetype>
            </d:prop>
            <d:status>HTTP/1.1 200 OK</d:status>
        </d:propstat>
    </d:response>
</d:multistatus>`)
}

func (h *CalDavHandler) GetCalendar(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	calendarID, err := uuid.Parse(c.Param("calendarID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid calendar id")
	}

	ical, err := h.u.ExportCalendarToICal(c.Request().Context(), userID, calendarID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Set("Content-Type", "text/calendar; charset=utf-8")
	return c.String(http.StatusOK, ical)
}

func (h *CalDavHandler) PutResource(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	calendarID, err := uuid.Parse(c.Param("calendarID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid calendar id")
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to read body")
	}

	if err := h.u.ImportFromICal(c.Request().Context(), userID, calendarID, string(body)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusCreated)
}

func (h *CalDavHandler) GetResource(c echo.Context) error {
	userID := c.Get("user_id").(uuid.UUID)
	resourceName := c.Param("resource")
	if !strings.HasSuffix(resourceName, ".ics") {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	icalUID := strings.TrimSuffix(resourceName, ".ics")

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
}
