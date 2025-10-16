package repository

import (
	"WEB/internal/app/ds"
	"context"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func (r *Repository) GetAllGases() ([]ds.Gas, error) {
	var gas []ds.Gas
	err := r.db.Find(&gas).Error
	if err != nil {
		return nil, err
	}
	return gas, nil
}

func (r *Repository) GetGasByID(id int) (*ds.Gas, error) {
	var gas ds.Gas
	err := r.db.First(&gas, id).Error
	if err != nil {
		return nil, err
	}
	return &gas, nil
}

func (r *Repository) SearchGasesByTitle(title string) ([]ds.Gas, error) {
	var gas []ds.Gas
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&gas).Error
	if err != nil {
		return nil, err
	}
	return gas, nil
}

// GasList returns gases with optional title search, excluding soft-deleted
func (r *Repository) GasList(search string) ([]ds.Gas, error) {
	var gases []ds.Gas
	q := r.db.Model(&ds.Gas{})
	if search != "" {
		q = q.Where("title ILIKE ?", "%"+search+"%")
	}
	if err := q.Find(&gases).Error; err != nil {
		return nil, err
	}
	return gases, nil
}

func (r *Repository) GasCreate(g *ds.Gas) error {
	return r.db.Create(g).Error
}

func (r *Repository) GasUpdate(id int, title *string, formula *string, molarMass *float64, description *string) error {
	updates := map[string]interface{}{}
	if title != nil {
		updates["title"] = *title
	}
	if formula != nil {
		updates["formula"] = *formula
	}
	if molarMass != nil {
		updates["molar_mass"] = *molarMass
	}
	if description != nil {
		updates["description"] = *description
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.Model(&ds.Gas{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) GasDelete(id int) error {
	// remove image from MinIO if present
	var gas ds.Gas
	if err := r.db.First(&gas, id).Error; err == nil && gas.ImageURL != "" {
		client, err := minio.New(r.minioEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(r.minioAccess, r.minioSecret, ""),
			Secure: r.minioUseSSL,
		})
		if err == nil {
			if parts := strings.Split(gas.ImageURL, "/"); len(parts) > 0 {
				oldName := parts[len(parts)-1]
				_ = client.RemoveObject(context.Background(), r.minioBucket, oldName, minio.RemoveObjectOptions{})
			}
		}
	}
	if err := r.db.Model(&ds.Gas{}).Where("id = ?", id).Update("image_url", "").Error; err != nil {
		return err
	}
	return r.db.Delete(&ds.Gas{}, id).Error
}

// stubbed MinIO upload; replace with real client later
func (r *Repository) GasUploadImage(ctx interface{ Done() <-chan struct{} }, id int, fileHeader *multipart.FileHeader) (string, error) {
	// init client
	client, err := minio.New(r.minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(r.minioAccess, r.minioSecret, ""),
		Secure: r.minioUseSSL,
	})
	if err != nil {
		return "", err
	}
	c := context.Background()
	// ensure bucket
	_ = client.MakeBucket(c, r.minioBucket, minio.MakeBucketOptions{})
	// normalize file name
	base := strings.ToLower(strings.ReplaceAll(fileHeader.Filename, " ", "-"))
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	ts := time.Now().Unix()
	objectName := name + "-" + strconv.FormatInt(ts, 10) + ext
	// open file
	f, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	// upload
	_, err = client.PutObject(c, r.minioBucket, objectName, f, fileHeader.Size, minio.PutObjectOptions{ContentType: fileHeader.Header.Get("Content-Type")})
	if err != nil {
		return "", err
	}
	url := r.minioBaseURL + "/" + r.minioBucket + "/" + objectName
	// remove previous image if exists
	var gas ds.Gas
	if err := r.db.First(&gas, id).Error; err == nil && gas.ImageURL != "" {
		// try delete old object
		if parts := strings.Split(gas.ImageURL, "/"); len(parts) > 0 {
			oldName := parts[len(parts)-1]
			_ = client.RemoveObject(c, r.minioBucket, oldName, minio.RemoveObjectOptions{})
		}
	}
	if err := r.db.Model(&ds.Gas{}).Where("id = ?", id).Update("image_url", url).Error; err != nil {
		return "", err
	}
	return url, nil
}

// FixedCreatorID returns singleton creator user id
func (r *Repository) FixedCreatorID() uint { return 1 }

// AddGasToDraft ensures a draft calculation and links gas (m-m)
func (r *Repository) AddGasToDraft(gasID uint, creatorID uint) error {
	// delegate to DB-backed implementation
	return r.addGasToDraftDB(gasID, creatorID)
}

// GetDraftCartInfo returns draft id and active gases count
func (r *Repository) GetDraftCartInfo(creatorID uint) (uint, int64, error) {
	return r.draftCartInfo(creatorID)
}
