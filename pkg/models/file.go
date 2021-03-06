package models

import (
	"encoder-backend/pkg/file"
	"github.com/ewanwalk/gorm"
	"os"
	"path/filepath"
	"time"
)

// TODO [integrity] on boot clear any "pending" or "running" encode status files

const (
	// Status codes
	FileStatusDeleted int64 = 0
	FileStatusEnabled int64 = 1
	// Encoder status codes
	FileEncodeStatusNotDone int64 = 0
	FileEncodeStatusDone    int64 = 1
	FileEncodeStatusPending int64 = 2
	FileEncodeStatusErrored int64 = 3
	FileEncodeStatusRunning int64 = 10
)

type File struct {
	ID            int64      `gorm:"AUTO_INCREMENT;primary_key" json:"id,omitempty"`
	PathID        int64      `gorm:"type:int(11);not null;index" json:"path_id,omitempty"`
	Name          string     `gorm:"type:varchar(255);not null" json:"name,omitempty"`
	Size          int64      `gorm:"type:bigint(20);not null;default:0" json:"size,omitempty"`
	Checksum      string     `gorm:"type:varchar(255);not null" json:"checksum,omitempty"`
	Source        string     `gorm:"type:varchar(512);not null" json:"source,omitempty"`
	Status        int64      `gorm:"type:int(2);default:1;index" json:"status"`
	StatusEncoder int64      `gorm:"type:int(2);default:0" json:"status_encoder"`
	CreatedAt     *time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at,omitempty"`
	UpdatedAt     *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at,omitempty"`

	// Relationships
	Encodes   []Encode   `json:"encodes,omitempty"`
	Path      *Path      `gorm:"association_autoupdate:false" json:"path,omitempty"`
	Revisions []Revision `json:"revisions,omitempty"`
}

// CurrentChecksum
// obtains the raw checksum by checking the file directly
func (f File) CurrentChecksum() (string, error) {
	return file.Checksum(filepath.Join(f.Source, f.Name))
}

// Exists
// checks the filesystem to ensure the file still exists
func (f File) Exists() bool {

	if !f.ExistsShallow() {
		return false
	}

	sum, err := f.CurrentChecksum()
	if err != nil || sum != f.Checksum {
		return false
	}

	return true
}

// ExistsShallow
// does a shallow check for file existence (only checking if the file exists on disk in the location known to us)
func (f File) ExistsShallow() bool {
	_, err := os.Stat(filepath.Join(f.Source, f.Name))
	if err != nil {
		return false
	}

	return true
}

func (f File) Filepath() string {
	return filepath.Join(f.Source, f.Name)
}

// Matches
// whether the source file matches the provided one
func (f File) Matches(file *File) bool {

	// if the checksum match we can pretty much guarantee they are the same file
	if f.Checksum == file.Checksum {
		return true
	}

	// if the names match we can probably assume that they are the same (?)
	if f.Name == file.Name {
		return true
	}

	return false
}

// FileNeedsEncode
// a gorm scope to find files which still need encoding
func FileNeedsEncode(db *gorm.DB) *gorm.DB {

	return db.Joins("left join paths on paths.id = files.path_id").
		Joins("left join encodes as e on e.file_id = files.id").
		Where("e.id = (?) OR e.id is null",
			db.Table("encodes").Select("MAX(id)").Where("file_id = files.id").QueryExpr(),
		).
		// This clause must match the models.PathAbleToEncode scope
		Where("paths.type in (?)", []int{
			PathTypeStandard,
		}).
		Where("files.size >= paths.minimum_file_size").
		Where("files.status = ?", FileStatusEnabled).
		Where(
			"((files.status_encoder = ? AND (files.checksum != e.checksum_at_end OR e.checksum_at_end is null)) OR e.status = ?)",
			FileEncodeStatusNotDone, EncodeCancelled,
		).
		Where("files.status_encoder not in (?)", []int64{
			FileEncodeStatusPending,
			FileEncodeStatusErrored,
		}).
		Order("paths.priority DESC")
}
