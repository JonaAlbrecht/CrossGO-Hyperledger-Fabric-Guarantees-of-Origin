package util

import (
	"fmt"
	"regexp"
	"strings"
)

// ADR-018 (v6.0): CEN-EN 16325 field validation.
// Validates GO fields against the European standard for energy guarantees of origin.

// validCountryCodes is the set of ISO 3166-1 alpha-2 codes for EU/EEA/CH member states
// that participate in the AIB GO framework.
var validCountryCodes = map[string]bool{
	"AT": true, "BE": true, "BG": true, "HR": true, "CY": true,
	"CZ": true, "DK": true, "EE": true, "FI": true, "FR": true,
	"DE": true, "GR": true, "HU": true, "IE": true, "IT": true,
	"LV": true, "LT": true, "LU": true, "MT": true, "NL": true,
	"PL": true, "PT": true, "RO": true, "SK": true, "SI": true,
	"ES": true, "SE": true, // EU-27
	"NO": true, "IS": true, "LI": true, // EEA
	"CH": true, "GB": true, // Associated
}

// validSupportSchemes lists the CEN-EN 16325 support scheme categories.
var validSupportSchemes = map[string]bool{
	"none":  true, // No support scheme
	"FIT":   true, // Feed-in tariff
	"FIP":   true, // Feed-in premium
	"quota": true, // Quota/obligation scheme
	"tax":   true, // Tax incentive
	"loan":  true, // Loan/investment aid
	"other": true, // Other support mechanism
}

// energySourcePattern matches EECS fact sheet energy source codes (e.g., "F01010100" = solar PV).
// Format: F + 8 digits, hierarchical from general to specific.
var energySourcePattern = regexp.MustCompile(`^F\d{8}$`)

// eicCodePattern matches European EIC codes for grid connection points.
// Format: 16 alphanumeric characters (ISO 6523 compliant).
var eicCodePattern = regexp.MustCompile(`^[A-Za-z0-9]{16}$`)

// ValidateCountryOfOrigin checks that the country code is a valid ISO 3166-1 alpha-2
// code for a country participating in the EU GO framework.
func ValidateCountryOfOrigin(country string) error {
	if country == "" {
		return nil // omitempty — not required
	}
	if !validCountryCodes[strings.ToUpper(country)] {
		return fmt.Errorf("invalid CountryOfOrigin %q: must be an EU/EEA ISO 3166-1 alpha-2 code", country)
	}
	return nil
}

// ValidateSupportScheme checks that the support scheme is a recognised CEN-EN 16325 category.
func ValidateSupportScheme(scheme string) error {
	if scheme == "" {
		return nil // omitempty
	}
	if !validSupportSchemes[strings.ToLower(scheme)] {
		return fmt.Errorf("invalid SupportScheme %q: must be one of none, FIT, FIP, quota, tax, loan, other", scheme)
	}
	return nil
}

// ValidateEnergySource checks that the energy source code matches the EECS fact sheet format.
func ValidateEnergySource(source string) error {
	if source == "" {
		return nil // omitempty
	}
	if !energySourcePattern.MatchString(source) {
		return fmt.Errorf("invalid EnergySource %q: must match EECS format F + 8 digits (e.g. F01010100)", source)
	}
	return nil
}

// ValidateGridConnectionPoint checks that the grid connection point matches EIC code format.
func ValidateGridConnectionPoint(eic string) error {
	if eic == "" {
		return nil // omitempty
	}
	if !eicCodePattern.MatchString(eic) {
		return fmt.Errorf("invalid GridConnectionPoint %q: must be a 16-character EIC code", eic)
	}
	return nil
}

// ValidateProductionPeriod checks that the production period timestamps are logically valid.
func ValidateProductionPeriod(start, end int64) error {
	if start == 0 && end == 0 {
		return nil // both omitted
	}
	if start != 0 && end == 0 {
		return fmt.Errorf("ProductionPeriodEnd is required when ProductionPeriodStart is set")
	}
	if start == 0 && end != 0 {
		return fmt.Errorf("ProductionPeriodStart is required when ProductionPeriodEnd is set")
	}
	if end <= start {
		return fmt.Errorf("ProductionPeriodEnd (%d) must be after ProductionPeriodStart (%d)", end, start)
	}
	// Max production period: 1 year (EN 16325 §6.3)
	const maxPeriodSeconds int64 = 366 * 24 * 3600
	if end-start > maxPeriodSeconds {
		return fmt.Errorf("production period exceeds maximum of 1 year (%d seconds)", end-start)
	}
	return nil
}

// ValidateCENFields runs all CEN-EN 16325 field validations on a GO's public fields.
func ValidateCENFields(country, gridConnection, supportScheme, energySource string, periodStart, periodEnd int64) error {
	if err := ValidateCountryOfOrigin(country); err != nil {
		return err
	}
	if err := ValidateGridConnectionPoint(gridConnection); err != nil {
		return err
	}
	if err := ValidateSupportScheme(supportScheme); err != nil {
		return err
	}
	if err := ValidateEnergySource(energySource); err != nil {
		return err
	}
	return ValidateProductionPeriod(periodStart, periodEnd)
}
