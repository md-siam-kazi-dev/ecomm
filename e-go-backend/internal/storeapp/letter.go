package storeapp

import (
	"bytes"
	"context"
	"html/template"

	"github.com/md-siam-kazi/e-go-backend/pkg/apperr"
)

// ApprovalLetter renders a printable approval document from trusted server data.
// PDF generation can replace this later without changing the API.
func (s *Service) ApprovalLetter(ctx context.Context, userID, id string, isAdmin bool) ([]byte, error) {
	app, err := s.repo.ApplicationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !isAdmin && app.ApplicantUserID != userID {
		return nil, errForbidden
	}
	if app.Status != StatusApproved {
		return nil, errNoLetter
	}

	approvedAt := app.UpdatedAt
	if app.ReviewedAt != nil {
		approvedAt = *app.ReviewedAt
	}

	data := ApprovalLetterData{
		PlatformName:      s.cfg.PlatformName,
		ApplicationNumber: app.ApplicationNumber,
		ApplicantName:     app.ApplicantName,
		StoreName:         app.StoreName,
		StoreCategory:     app.Category,
		ApprovalDate:      approvedAt.UTC().Format("2 January 2006"),
		FeeEnabled:        s.cfg.FeeEnabled,
		FeeAmount:         s.cfg.FeeAmount,
		FeeCurrency:       s.cfg.FeeCurrency,
	}

	var buf bytes.Buffer
	if err := approvalLetterTemplate.Execute(&buf, data); err != nil {
		return nil, apperr.ErrInternal.WithCause(err)
	}
	return buf.Bytes(), nil
}

var approvalLetterTemplate = template.Must(template.New("approval-letter").Parse(`
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>Store Approval Letter — {{.ApplicationNumber}}</title>
<style>
  :root { color-scheme: light; }
  body { font-family: ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, sans-serif; margin: 0; background: #f8fafc; color: #0f172a; }
  .sheet { max-width: 760px; margin: 32px auto; background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; padding: 48px; }
  h1 { font-size: 22px; margin: 0 0 4px; }
  .muted { color: #64748b; font-size: 13px; }
  .brand { font-weight: 700; letter-spacing: .02em; color: #2563eb; }
  table { width: 100%; border-collapse: collapse; margin: 24px 0; font-size: 14px; }
  th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid #e2e8f0; vertical-align: top; }
  th { width: 38%; color: #64748b; font-weight: 600; }
  .notice { background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 10px; padding: 16px; font-size: 14px; margin-top: 8px; }
  .foot { margin-top: 32px; font-size: 12px; color: #94a3b8; }
  @media print { body { background: #fff; } .sheet { border: none; margin: 0; padding: 0; } }
</style>
</head>
<body>
  <div class="sheet">
    <div class="brand">{{.PlatformName}}</div>
    <h1>Store Application Approval Letter</h1>
    <p class="muted">Document reference: {{.ApplicationNumber}}-L · Issued {{.ApprovalDate}}</p>

    <p>Dear {{.ApplicantName}},</p>
    <p>
      We are pleased to inform you that your application to open a store on
      {{.PlatformName}} has been <strong>approved</strong>.
    </p>

    <table>
      <tr><th>Application number</th><td>{{.ApplicationNumber}}</td></tr>
      <tr><th>Store name</th><td>{{.StoreName}}</td></tr>
      <tr><th>Store category</th><td>{{.StoreCategory}}</td></tr>
      <tr><th>Approval date</th><td>{{.ApprovalDate}}</td></tr>
      <tr><th>Application fee</th><td>{{if .FeeEnabled}}{{.FeeAmount}} {{.FeeCurrency}}{{else}}No application fee is required at this time.{{end}}</td></tr>
    </table>

    <div class="notice">
      <strong>Important:</strong> this approval does not by itself activate your
      store or grant dashboard access. Your store becomes eligible for activation
      only after {{if .FeeEnabled}}the application fee of {{.FeeAmount}} {{.FeeCurrency}} is paid and verified on the server,{{end}}
      and any other required checks are completed.
    </div>

    <p class="muted" style="margin-top:24px">
      Conditions of approval: the information you submitted must remain accurate and
      complete; you must comply with the platform's vendor terms and store policies;
      and you must keep your store details and verification information up to date.
    </p>

    <div class="foot">
      This letter is generated from {{.PlatformName}} records and is valid without a
      signature. Letter status: <strong>valid</strong>.
    </div>
  </div>
</body>
</html>`))
