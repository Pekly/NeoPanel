package api

import (
	"net/http"
	"strings"

	"github.com/Pekly/NeoPanel/internal/runner"
)

func importAdmin(r *http.Request) bool {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	return claims != nil && claims.Role == "admin"
}

// Retain diagnostics for operators without sending host paths to server users.
func importJobForViewer(r *http.Request, job sftpImportJob) sftpImportJob {
	switch job.State {
	case "completed":
		job.UserMessage = "Files imported. Review your server settings before starting."
	case "cancelled":
		job.UserMessage = "Import cancelled or timed out. No files were imported."
	case "failed":
		job.UserMessage = importUserError(job.Error)
	default:
		if strings.HasPrefix(job.Message, "Cancellation requested") {
			job.UserMessage = "Cancelling your import…"
		}
	}
	if !importAdmin(r) {
		job.SafetyCopy = ""
		job.Error = ""
		job.Message = job.UserMessage
	}
	return job
}

func importErrorForViewer(r *http.Request, message string) string {
	if importAdmin(r) {
		return message
	}
	return importUserError(message)
}

func importUserError(message string) string {
	switch {
	case strings.Contains(message, runner.ErrRestoreRecovery.Error()):
		return "The import could not finish safely. Contact support before changing files or trying again."
	case strings.Contains(message, runner.ErrRestoreRetention.Error()):
		return "Imports are temporarily unavailable for this server. Contact support."
	case strings.Contains(message, runner.ErrImportTargetNotEmpty.Error()):
		return "Your server still has files. Back up and move or remove them before importing."
	case strings.Contains(message, runner.ErrServerNotStopped.Error()):
		return "Stop your server before importing files."
	case strings.Contains(message, runner.ErrServerOperation.Error()):
		return "Another server action is in progress. Try again shortly."
	case strings.Contains(message, "host key mismatch"):
		return "The host fingerprint did not match. Verify it with your provider before trying again."
	case strings.Contains(message, "outbound address denied"):
		return "This address is not available for imports. Contact support if you need to use it."
	case strings.Contains(message, "authentication failed"):
		return "Could not sign in to the source. Check its address and login details."
	case strings.Contains(message, "invalid private key"):
		return "Check your private key and passphrase. Use a PEM or OpenSSH key."
	case strings.Contains(message, "host resolution failed"), strings.Contains(message, "connection failed"):
		return "Could not connect to the source. Check its address and port, then try again."
	case strings.Contains(message, "host fingerprint required"):
		return "Enter a valid SHA256 fingerprint and verify it with your provider."
	case strings.Contains(message, "invalid byte limit"):
		return "Choose a transfer size within the allowed range."
	case strings.Contains(message, "byte limit"):
		return "The selected files exceed the max copy size limit. Increase Max copy size (GiB) in Advanced options or copy fewer files."
	case strings.Contains(message, "remote entry changed"), strings.Contains(message, "remote file changed"), (strings.Contains(message, "remote") && strings.Contains(message, " changed")):
		return "Source files changed during the copy. Stop the source server and try again."
	case strings.Contains(message, "entry limit"), strings.Contains(message, "depth limit"), strings.Contains(message, "listing failed or exceeded limits"):
		return "This selection is too large to list. Choose a smaller source folder or contact support."
	case strings.Contains(message, "symlink"), strings.Contains(message, "unsafe remote"), strings.Contains(message, "path collision"):
		return "Some source files are unsupported. Check for symbolic links or conflicting filenames."
	case strings.Contains(message, "invalid host"):
		return "Enter a valid SFTP address and port."
	case strings.Contains(message, "invalid remote"), strings.Contains(message, "remote root unavailable"):
		return "Could not open the source folder. Check its path and access permissions."
	case strings.Contains(message, "invalid username"):
		return "Enter a valid source username."
	case strings.Contains(message, "credential too long"):
		return "The login details exceed the allowed size. Check the password or key."
	case strings.Contains(message, "provide exactly one"):
		return "Choose either a password or a private key to connect."
	case strings.Contains(message, "path filter"):
		return "Check the selected and excluded paths. Use relative paths without wildcards."
	case strings.Contains(message, "invalid bandwidth"):
		return "Enter a valid bandwidth limit, or 0 for unlimited."
	case strings.Contains(message, "invalid timeout"):
		return "Choose a timeout between 1 and 240 minutes."
	case strings.Contains(message, "permission was revoked"):
		return "Your access changed during the import. Sign in again or contact the server owner."
	default:
		return "The import could not finish. Check your connection details, or contact support if the problem continues."
	}
}
