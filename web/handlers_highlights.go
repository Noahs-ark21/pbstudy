package web

import (
	"net/http"

	"github.com/rohanthewiz/rweb"

	"github.com/rohanthewiz/pbstudy/bible"
	"github.com/rohanthewiz/pbstudy/study"
)

// handleHighlight colours a verse, or clears its colour when the submitted
// colour is "none". The verse hub's swatches are submit buttons on one form,
// so the colour arrives as the name of whichever button was pressed.
func (s *Server) handleHighlight(ctx rweb.Context) error {
	req := ctx.Request()
	back := safeReturn(req.FormValue("return"), "/")

	at, err := bible.ParseRef(req.FormValue("at"))
	if err != nil || at.IsChapter() {
		return s.renderError(ctx, http.StatusBadRequest, "Cannot highlight that",
			"A highlight has to be on a specific verse.", err)
	}

	color := req.FormValue("color")
	switch {
	case color == "none":
		err = study.ClearHighlight(s.store.Study, at)
	case study.IsHighlightColor(color):
		err = study.SetHighlight(s.store.Study, at, color)
	default:
		return s.renderError(ctx, http.StatusBadRequest, "Cannot highlight that",
			"There is no highlight colour called \""+color+"\".", nil)
	}
	if err != nil {
		return s.renderError(ctx, http.StatusInternalServerError,
			"Cannot save highlight", "The study database could not be written.", err)
	}

	return ctx.Redirect(http.StatusSeeOther, back)
}
