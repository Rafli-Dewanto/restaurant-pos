package model

type WishListResponse struct {
	Cakes []CakeModel `json:"cakes"`
}

type AddToWishListRequest struct {
	CakeID int64 `json:"cake_id" validate:"required"`
}
