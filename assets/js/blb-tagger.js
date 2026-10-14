// Blue Letter Bible ScriptTagger configuration.
//
// Ordering is the whole point of this file. The tagger REPLACES window.BLB.Tagger
// with its own defaults when it executes (NKJV popups, no hyperlinks, and only
// its built-in "noTag doNotTag" exclusions), so a config object assigned
// before the script tag is thrown away. BLB's documented embed assigns
// individual BLB.Tagger properties after the script instead, and that is what
// this does:
//
//   <script src=".../BLB_ScriptTagger-min.js" defer>   1. tagger executes
//   <script src="/js/blb-tagger.js" defer data-…>      2. this file overrides
//   <script src="/js/app.js" defer>                     3. app behaviour
//
// Deferred scripts run in document order, so step 2 always follows step 1.
// The tagging pass itself runs later still: the tagger first loads its own
// dependencies (YUI, BLB's loader) asynchronously, and it parses
// NoSearchClassNames into its exclusion set at that point, so properties set
// here are in place before any reference on the page is tagged. Verified in
// Chrome on 2026-10-14: with this order popups quote the configured version,
// and a "blb-no-tag" container is left alone; with the old order both failed.
//
// Values come from data attributes on this script's own tag so the server
// decides them per page (the translation being read) without an inline
// script. If the tagger failed to load (offline), window.BLB is undefined and
// this does nothing.
(function () {
  "use strict";

  var me = document.currentScript;
  if (!me || !window.BLB || !window.BLB.Tagger) {
    return;
  }

  var tagger = window.BLB.Tagger;
  var translation = me.getAttribute("data-translation");
  if (translation) {
    tagger.Translation = translation;
  }

  // "all" links each tagged reference to its BLB page as well as giving it a
  // popup. The tagger's values are strings ("all" / "none" / "hover"), which
  // is why the old boolean `HyperLinks: true` never took effect.
  tagger.HyperLinks = "all";
  tagger.TargetNewWindow = true;

  // Keep the tagger's own exclusions and add ours. A space-separated class
  // list, parsed by the tagger when it starts tagging.
  var noTag = me.getAttribute("data-no-tag");
  if (noTag) {
    tagger.NoSearchClassNames = "noTag doNotTag " + noTag;
  }
})();
