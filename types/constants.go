package types

import "time"

var (
	EXP_REFRESH_TOKEN = 7 * 24 * time.Hour // 7 days in hours
)

var (
	USERNAME_REGEX = "^[a-z][a-z0-9]{0,29}$"
	PASSWORD_REGEX = "^.{6,29}$"
)

var (
	LABEL_REPLACEMENT = "vật tư thay thế"
	LABEL_CONSUMABLE  = "vật tư tiêu hao"
)

// Material Management Types

var (
	SECTOR_MECHANICAL           = "Cơ khí"
	SECTOR_WEAPONS              = "Vũ khí"
	SECTOR_HULL                 = "Vỏ Tàu"
	SECTOR_DOCK                 = "Đà đốc"
	SECTOR_ELECTRONICS          = "Điện tàu"
	SECTOR_PROPULSION           = "Động lực"
	SECTOR_VALVE_PIPE           = "Van ống"
	SECTOR_ELECTRONICS_TACTICAL = "KT-ĐT"
	SECTOR_DECORATIVE           = "Trang trí"
	SECTOR_ELECTRICAL           = "Cơ điện"
)

var (
	SECTOR_LIST = []string{
		SECTOR_MECHANICAL,
		SECTOR_WEAPONS,
		SECTOR_HULL,
		SECTOR_DOCK,
		SECTOR_ELECTRONICS,
		SECTOR_PROPULSION,
		SECTOR_VALVE_PIPE,
		SECTOR_ELECTRONICS_TACTICAL,
		SECTOR_DECORATIVE,
		SECTOR_ELECTRICAL,
	}
)

var ShortSectorList = map[string]string{
	SECTOR_MECHANICAL:           "CK",
	SECTOR_WEAPONS:              "VK",
	SECTOR_HULL:                 "VT",
	SECTOR_DOCK:                 "ĐĐ",
	SECTOR_ELECTRONICS:          "ĐT",
	SECTOR_PROPULSION:           "ĐL",
	SECTOR_VALVE_PIPE:           "VỐ",
	SECTOR_ELECTRONICS_TACTICAL: "KT",
	SECTOR_DECORATIVE:           "TT",
	SECTOR_ELECTRICAL:           "CĐ",
}

var (
	MAINTENANCE_TIER_DOCK   = "SCCĐ"
	MAINTENANCE_TIER_SMALL  = "SCCN"
	MAINTENANCE_TIER_MEDIUM = "SCCV"
)

var (
	MAINTENANCE_TIER_LIST = []string{
		MAINTENANCE_TIER_DOCK,
		MAINTENANCE_TIER_SMALL,
		MAINTENANCE_TIER_MEDIUM,
	}
)

var (
	MATERIALS_REQUEST_PREFIX = "YCVT-"
)
