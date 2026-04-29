package scorer

// Scoring rubric thresholds — all numeric cut-points that encode rubric decisions live here.
// Update this file when the rubric changes; do not scatter literals across dimension files.

const (
	// D3 — Anti-Pattern Quality
	d3StrongMarkersHigh = 8 // StrongMarkers above this → max directive-language score
	d3StrongMarkersMid  = 4 // StrongMarkers above this → mid directive-language score
	d3NeverCountMin     = 3 // fallback: NEVER occurrences above this → capped score
	d3AntiInstrHigh     = 5 // ≥ this many anti-pattern instructions → bonus +2
	d3AntiInstrMid      = 3 // ≥ this many anti-pattern instructions → bonus +1
	d3Max               = 15

	// D4 — Specification Compliance
	d4DescLenMid     = 100 // description bytes above this → +2
	d4DescLenHigh    = 200 // description bytes above this → +1 additional
	d4AndOrCountHigh = 3   // and/or conjunctions above this → -2 (over-stuffed)
	d4AndOrCountMid  = 1   // and/or conjunctions above this → -1
	d4Max            = 15
	d4MaxWithBonus   = 17
	d4PenaltyCap     = 2 // max penalty per dir scan

	// D5 — Progressive Disclosure (token path)
	d5TokenCompact  = 800  // < this → full score with refs
	d5TokenModerate = 1200 // < this → tier-2 score
	d5TokenVerbose  = 1600 // < this → tier-3 score
	d5TokenLong     = 2400 // < this → tier-2 score without refs
	d5TokenVeryLong = 4000 // < this → tier-3 score without refs

	// D5 — Progressive Disclosure (line path fallback)
	d5LinesCompact  = 100 // < this → full score with refs
	d5LinesModerate = 150 // < this → tier-2 score
	d5LinesVerbose  = 200 // < this → tier-3 score
	d5LinesLong     = 300 // < this → tier-2 score without refs
	d5LinesVeryLong = 500 // < this → tier-3 score without refs

	// D8 — Practical Usability
	d8BlocksHigh = 5 // code blocks above this → max block delta
	d8BlocksMid  = 2 // code blocks above this → mid block delta
	d8BaseScore  = 5
	d8Max        = 15

	// D9 — Eval Validation
	d9CoverageMin    = 80 // coverage % must meet or exceed this to score full summary points
	d9ScenariosHigh  = 3  // ≥ this many valid scenarios → full scenario score
	d9ScenariosMid   = 1  // ≥ this many valid scenarios → partial scenario score
	d9EvalsDirPoints = 3  // points awarded simply for having an evals/ directory
	d9Max            = 20
)
