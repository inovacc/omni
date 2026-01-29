// Package brdoc provides Brazilian document validation and generation.
// Wraps github.com/inovacc/brdoc for CPF and CNPJ operations.
package brdoc

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/inovacc/brdoc"
)

// Options configures brdoc command behavior
type Options struct {
	Generate bool // Generate a new document
	Validate bool // Validate a document
	Format   bool // Format a document
	Count    int  // Number of documents to generate
	Legacy   bool // Use legacy numeric-only CNPJ format
	JSON     bool // Output as JSON
}

// CPFResult represents CPF operation result
type CPFResult struct {
	CPF   string `json:"cpf"`
	Valid bool   `json:"valid,omitempty"`
	Error string `json:"error,omitempty"`
	State string `json:"state,omitempty"`
}

// CNPJResult represents CNPJ operation result
type CNPJResult struct {
	CNPJ  string `json:"cnpj"`
	Valid bool   `json:"valid,omitempty"`
	Error string `json:"error,omitempty"`
}

// CPFListResult represents multiple CPF results
type CPFListResult struct {
	Count int         `json:"count"`
	CPFs  []CPFResult `json:"cpfs"`
}

// CNPJListResult represents multiple CNPJ results
type CNPJListResult struct {
	Count int          `json:"count"`
	CNPJs []CNPJResult `json:"cnpjs"`
}

var (
	cpfHandler  = brdoc.NewCPF()
	cnpjHandler = brdoc.NewCNPJ()
)

// RunCPF executes CPF operations
func RunCPF(w io.Writer, args []string, opts Options) error {
	if opts.Generate {
		return generateCPF(w, opts)
	}

	if opts.Validate {
		return validateCPF(w, args, opts)
	}

	if opts.Format {
		return formatCPF(w, args, opts)
	}

	// Default: generate one
	opts.Count = 1

	return generateCPF(w, opts)
}

// RunCNPJ executes CNPJ operations
func RunCNPJ(w io.Writer, args []string, opts Options) error {
	if opts.Generate {
		return generateCNPJ(w, opts)
	}

	if opts.Validate {
		return validateCNPJ(w, args, opts)
	}

	if opts.Format {
		return formatCNPJ(w, args, opts)
	}

	// Default: generate one
	opts.Count = 1

	return generateCNPJ(w, opts)
}

func generateCPF(w io.Writer, opts Options) error {
	count := opts.Count
	if count <= 0 {
		count = 1
	}

	if opts.JSON {
		result := CPFListResult{Count: count}
		for i := 0; i < count; i++ {
			cpf := cpfHandler.Generate()
			formatted, _ := cpfHandler.Format(cpf)
			state := cpfHandler.CheckOrigin(cpf)
			result.CPFs = append(result.CPFs, CPFResult{
				CPF:   formatted,
				State: state,
			})
		}

		return json.NewEncoder(w).Encode(result)
	}

	for i := 0; i < count; i++ {
		cpf := cpfHandler.Generate()
		formatted, _ := cpfHandler.Format(cpf)
		_, _ = fmt.Fprintln(w, formatted)
	}

	return nil
}

func validateCPF(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 {
		return fmt.Errorf("cpf: no document provided")
	}

	if opts.JSON {
		var results []CPFResult

		for _, arg := range args {
			result := CPFResult{CPF: arg}
			if cpfHandler.Validate(arg) {
				result.Valid = true
				result.State = cpfHandler.CheckOrigin(arg)
			} else {
				result.Valid = false
				result.Error = "invalid CPF"
			}

			results = append(results, result)
		}

		if len(results) == 1 {
			return json.NewEncoder(w).Encode(results[0])
		}

		return json.NewEncoder(w).Encode(CPFListResult{Count: len(results), CPFs: results})
	}

	allValid := true

	for _, arg := range args {
		if cpfHandler.Validate(arg) {
			state := cpfHandler.CheckOrigin(arg)
			_, _ = fmt.Fprintf(w, "%s: valid (state: %s)\n", arg, state)
		} else {
			_, _ = fmt.Fprintf(w, "%s: invalid\n", arg)
			allValid = false
		}
	}

	if !allValid {
		return fmt.Errorf("one or more CPFs are invalid")
	}

	return nil
}

func formatCPF(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 {
		return fmt.Errorf("cpf: no document provided")
	}

	if opts.JSON {
		var results []CPFResult

		for _, arg := range args {
			clean := cleanDoc(arg)
			formatted, _ := cpfHandler.Format(clean)
			results = append(results, CPFResult{CPF: formatted})
		}

		if len(results) == 1 {
			return json.NewEncoder(w).Encode(results[0])
		}

		return json.NewEncoder(w).Encode(CPFListResult{Count: len(results), CPFs: results})
	}

	for _, arg := range args {
		clean := cleanDoc(arg)
		formatted, _ := cpfHandler.Format(clean)
		_, _ = fmt.Fprintln(w, formatted)
	}

	return nil
}

func generateCNPJ(w io.Writer, opts Options) error {
	count := opts.Count
	if count <= 0 {
		count = 1
	}

	if opts.JSON {
		result := CNPJListResult{Count: count}
		for i := 0; i < count; i++ {
			var cnpj string
			if opts.Legacy {
				cnpj = cnpjHandler.GenerateLegacy()
			} else {
				cnpj = cnpjHandler.Generate()
			}

			formatted, _ := cnpjHandler.Format(cnpj)
			result.CNPJs = append(result.CNPJs, CNPJResult{
				CNPJ: formatted,
			})
		}

		return json.NewEncoder(w).Encode(result)
	}

	for i := 0; i < count; i++ {
		var cnpj string
		if opts.Legacy {
			cnpj = cnpjHandler.GenerateLegacy()
		} else {
			cnpj = cnpjHandler.Generate()
		}

		formatted, _ := cnpjHandler.Format(cnpj)
		_, _ = fmt.Fprintln(w, formatted)
	}

	return nil
}

func validateCNPJ(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 {
		return fmt.Errorf("cnpj: no document provided")
	}

	if opts.JSON {
		var results []CNPJResult

		for _, arg := range args {
			result := CNPJResult{CNPJ: arg}
			if cnpjHandler.Validate(arg) {
				result.Valid = true
			} else {
				result.Valid = false
				result.Error = "invalid CNPJ"
			}

			results = append(results, result)
		}

		if len(results) == 1 {
			return json.NewEncoder(w).Encode(results[0])
		}

		return json.NewEncoder(w).Encode(CNPJListResult{Count: len(results), CNPJs: results})
	}

	allValid := true

	for _, arg := range args {
		if cnpjHandler.Validate(arg) {
			_, _ = fmt.Fprintf(w, "%s: valid\n", arg)
		} else {
			_, _ = fmt.Fprintf(w, "%s: invalid\n", arg)
			allValid = false
		}
	}

	if !allValid {
		return fmt.Errorf("one or more CNPJs are invalid")
	}

	return nil
}

func formatCNPJ(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 {
		return fmt.Errorf("cnpj: no document provided")
	}

	if opts.JSON {
		var results []CNPJResult

		for _, arg := range args {
			clean := cleanDoc(arg)
			formatted, _ := cnpjHandler.Format(clean)
			results = append(results, CNPJResult{CNPJ: formatted})
		}

		if len(results) == 1 {
			return json.NewEncoder(w).Encode(results[0])
		}

		return json.NewEncoder(w).Encode(CNPJListResult{Count: len(results), CNPJs: results})
	}

	for _, arg := range args {
		clean := cleanDoc(arg)
		formatted, _ := cnpjHandler.Format(clean)
		_, _ = fmt.Fprintln(w, formatted)
	}

	return nil
}

// cleanDoc removes formatting characters from a document
func cleanDoc(doc string) string {
	doc = strings.ReplaceAll(doc, ".", "")
	doc = strings.ReplaceAll(doc, "-", "")
	doc = strings.ReplaceAll(doc, "/", "")
	doc = strings.ReplaceAll(doc, " ", "")

	return doc
}

// Library functions for direct use

// GenerateCPF generates a valid CPF
func GenerateCPF() string {
	return cpfHandler.Generate()
}

// GenerateCPFFormatted generates a formatted valid CPF
func GenerateCPFFormatted() string {
	formatted, _ := cpfHandler.Format(cpfHandler.Generate())
	return formatted
}

// ValidateCPF validates a CPF
func ValidateCPF(cpf string) bool {
	return cpfHandler.Validate(cpf)
}

// FormatCPF formats a CPF as XXX.XXX.XXX-XX
func FormatCPF(cpf string) string {
	formatted, _ := cpfHandler.Format(cleanDoc(cpf))
	return formatted
}

// CPFState returns the state/region for a CPF
func CPFState(cpf string) string {
	return cpfHandler.CheckOrigin(cpf)
}

// GenerateCNPJ generates a valid alphanumeric CNPJ
func GenerateCNPJ() string {
	return cnpjHandler.Generate()
}

// GenerateCNPJLegacy generates a valid numeric-only CNPJ
func GenerateCNPJLegacy() string {
	return cnpjHandler.GenerateLegacy()
}

// GenerateCNPJFormatted generates a formatted valid CNPJ
func GenerateCNPJFormatted() string {
	formatted, _ := cnpjHandler.Format(cnpjHandler.Generate())
	return formatted
}

// ValidateCNPJ validates a CNPJ (supports alphanumeric)
func ValidateCNPJ(cnpj string) bool {
	return cnpjHandler.Validate(cnpj)
}

// FormatCNPJ formats a CNPJ as XX.XXX.XXX/XXXX-XX
func FormatCNPJ(cnpj string) string {
	formatted, _ := cnpjHandler.Format(cleanDoc(cnpj))
	return formatted
}
