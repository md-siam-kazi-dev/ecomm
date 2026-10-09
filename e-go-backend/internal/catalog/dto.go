package catalog

import (
	"sort"
	"strings"
	"time"
)

// isFeaturedRating is the rating at or above which a product is treated as
// featured (there is no curated flag in the schema yet).
const isFeaturedRating = 4.0

// SpecificationDTO is a label/value spec derived from products.attributes.
type SpecificationDTO struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// ProductDTO is a storefront product.
type ProductDTO struct {
	ID               string             `json:"id"`
	Slug             string             `json:"slug"`
	Name             string             `json:"name"`
	ShortDescription string             `json:"short_description"`
	Description      string             `json:"description"`
	Price            float64            `json:"price"`
	CompareAtPrice   *float64           `json:"compare_at_price,omitempty"`
	SKU              string             `json:"sku"`
	Stock            int64              `json:"stock"`
	ImageURL         string             `json:"image_url,omitempty"`
	ImageAlt         string             `json:"image_alt,omitempty"`
	Category         string             `json:"category"`
	CategorySlug     string             `json:"category_slug"`
	Brand            string             `json:"brand,omitempty"`
	Rating           float64            `json:"rating"`
	ReviewCount      int64              `json:"review_count"`
	IsFeatured       bool               `json:"is_featured"`
	IsNew            bool               `json:"is_new"`
	IsBestSeller     bool               `json:"is_best_seller"`
	Specifications   []SpecificationDTO `json:"specifications"`
	SeoTitle         string             `json:"seo_title,omitempty"`
	SeoDescription   string             `json:"seo_description,omitempty"`
	CreatedAt        string             `json:"created_at"`
	UpdatedAt        string             `json:"updated_at"`
}

// CategoryDTO is a storefront category.
type CategoryDTO struct {
	ID           string `json:"id"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	ImageURL     string `json:"image_url,omitempty"`
	ImageAlt     string `json:"image_alt,omitempty"`
	ProductCount int64  `json:"product_count"`
}

// ReviewDTO is a published product review.
type ReviewDTO struct {
	ID                 string `json:"id"`
	ProductID          string `json:"product_id"`
	ProductName        string `json:"product_name,omitempty"`
	ProductSlug        string `json:"product_slug,omitempty"`
	Author             string `json:"author"`
	Rating             int    `json:"rating"`
	Title              string `json:"title"`
	Comment            string `json:"comment"`
	IsVerifiedPurchase bool   `json:"is_verified_purchase"`
	HelpfulCount       int64  `json:"helpful_count"`
	CreatedAt          string `json:"created_at"`
}

func toProductDTO(r ProductRecord) ProductDTO {
	return ProductDTO{
		ID:               r.ID,
		Slug:             r.Slug,
		Name:             r.Name,
		ShortDescription: r.ShortDescription,
		Description:      r.Description,
		Price:            r.Price,
		CompareAtPrice:   r.CompareAtPrice,
		SKU:              r.SKU,
		Stock:            r.Stock,
		ImageURL:         r.ImageURL,
		ImageAlt:         r.ImageAlt,
		Category:         r.Category,
		CategorySlug:     r.CategorySlug,
		Brand:            r.Brand,
		Rating:           r.Rating,
		ReviewCount:      r.ReviewCount,
		IsFeatured:       r.Rating >= isFeaturedRating,
		IsNew:            r.IsNew,
		IsBestSeller:     r.SalesUnits > 0,
		Specifications:   toSpecifications(r.Attributes),
		SeoTitle:         r.SeoTitle,
		SeoDescription:   r.SeoDescription,
		CreatedAt:        r.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        r.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toSpecifications(attrs map[string]string) []SpecificationDTO {
	if len(attrs) == 0 {
		return []SpecificationDTO{}
	}
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	specs := make([]SpecificationDTO, 0, len(keys))
	for _, k := range keys {
		v := strings.TrimSpace(attrs[k])
		if v == "" {
			continue
		}
		specs = append(specs, SpecificationDTO{Label: humanizeKey(k), Value: v})
	}
	return specs
}

func humanizeKey(key string) string {
	key = strings.ReplaceAll(key, "_", " ")
	key = strings.Join(strings.Fields(key), " ")
	if key == "" {
		return key
	}
	return strings.ToUpper(key[:1]) + key[1:]
}

func toCategoryDTO(r CategoryRecord) CategoryDTO {
	return CategoryDTO{
		ID:           r.ID,
		Slug:         r.Slug,
		Name:         r.Name,
		Description:  r.Description,
		ImageURL:     r.ImageURL,
		ImageAlt:     r.Name + " category",
		ProductCount: r.ProductCount,
	}
}

func toReviewDTO(r ReviewRecord) ReviewDTO {
	return ReviewDTO{
		ID:                 r.ID,
		ProductID:          r.ProductID,
		ProductName:        r.ProductName,
		ProductSlug:        r.ProductSlug,
		Author:             r.Author,
		Rating:             r.Rating,
		Title:              r.Title,
		Comment:            r.Comment,
		IsVerifiedPurchase: r.VerifiedPurchase,
		HelpfulCount:       r.HelpfulCount,
		CreatedAt:          r.CreatedAt.UTC().Format(time.RFC3339),
	}
}
