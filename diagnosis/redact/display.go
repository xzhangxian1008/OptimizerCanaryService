package redact

// DisplaySQL, DisplayPlan and DisplayBinding fail closed for report rendering.
func DisplaySQL(text string) string {
	value, err := SQL(text)
	if err != nil {
		return Omitted
	}
	return value
}

func DisplayPlan(text string) string {
	value, err := Plan(text)
	if err != nil {
		return Omitted
	}
	return value
}

func DisplayPlanWithSQL(text, sqlText, schema string) string {
	value, err := PlanWithSQL(text, sqlText, schema)
	if err != nil {
		return Omitted
	}
	return value
}

// DisplayAlignedPlanWithSQL formats a redacted plan for a monospace report.
func DisplayAlignedPlanWithSQL(text, sqlText, schema string) string {
	return alignPlanColumns(DisplayPlanWithSQL(text, sqlText, schema))
}

func DisplayBinding(original, hinted string) string {
	value, err := Binding(original, hinted)
	if err != nil {
		return Omitted
	}
	return value
}
