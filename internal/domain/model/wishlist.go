package model

type WishListResponse struct {
	Menus []MenuModel `json:"menus"`
}

type AddToWishListRequest struct {
	MenuID int64 `json:"menu_id" validate:"required"`
}
