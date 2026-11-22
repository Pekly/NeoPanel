package sftp

import "github.com/pkg/sftp"

// Logical sizes (including holes) are capped even without a disk assignment.
// This is a per-file bound, not an aggregate filesystem quota.
const maxSFTPFileSize int64 = 8 << 30

// sizeLimit re-reads the assignment for every mutation, including old handles.
func (h *chrootHandler) sizeLimit() (int64, error) {
	srv, err := h.server.db.GetServer(h.serverID)
	if err != nil || srv == nil {
		return 0, sftp.ErrSSHFxPermissionDenied
	}
	limit := maxSFTPFileSize
	// Compare before multiplying to avoid overflow for large assignments.
	if srv.DiskLimitMB > 0 && srv.DiskLimitMB < maxSFTPFileSize/(1<<20) {
		limit = srv.DiskLimitMB * (1 << 20)
	}
	return limit, nil
}

func checkedWriteEnd(off int64, length int, limit int64) (int64, error) {
	if off < 0 || off > limit || int64(length) > limit-off {
		return 0, sftp.ErrSSHFxFailure
	}
	return off + int64(length), nil
}
