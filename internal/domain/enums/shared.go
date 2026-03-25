package enums

type UserType string

const (
	User  UserType = "USER"
	Admin UserType = "ADMIN"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
	UserStatusPending   UserStatus = "PENDING"
	UserStatusInactive  UserStatus = "INACTIVE"
)

type ContactCategory string

const (
	CategoryWebsiteDevelopment ContactCategory = "website_development"
	CategoryFeedback           ContactCategory = "online_tools_feedback"
	CategoryGeneralEnquiry     ContactCategory = "general_enquiry"
	CategoryPartnership        ContactCategory = "partnership"
)

type ActionType string

const (
	ActionConvert           ActionType = "convert"
	ActionOptimize          ActionType = "optimize"
	ActionAnalyze           ActionType = "analyze"
	ActionSVGOptimize       ActionType = "svg_optimize"
	ActionCrop              ActionType = "crop"
	ActionResize            ActionType = "resize"
	ActionToBase64          ActionType = "to_base64"
	ActionFromBase64        ActionType = "from_base64"
	ActionFaviconGenerate   ActionType = "favicon_generate"
	ActionOptimizeTwitter   ActionType = "optimize_twitter"
	ActionOptimizeWhatsApp  ActionType = "optimize_whatsapp"
	ActionOptimizeWeb       ActionType = "optimize_web"
	ActionOptimizeCustom    ActionType = "optimize_custom"
	ActionOptimizeInstagram ActionType = "optimize_instagram"
	ActionOptimizeYouTube   ActionType = "optimize_youtube"
	ActionOptimizeSEO       ActionType = "optimize_seo"

	ActionPDFCompress    ActionType = "pdf_compress"
	ActionPDFCompressPro ActionType = "pdf_compress_pro"

	ActionImageToPDF ActionType = "image_to_pdf"
	ActionPDFToImage ActionType = "pdf_to_image"

	ActionExifScrubber ActionType = "exif_scrubber"

	ActionHEICToImage ActionType = "heic_to_image"
	ActionImageToHEIC ActionType = "image_to_heic"

	ActionImageColorEffects ActionType = "image_color_effects"
	ActionImageDPIChanger   ActionType = "image_dpi_changer"
	ActionImageDPIChecker   ActionType = "image_dpi_checker"

	ActionPasswordGenerator ActionType = "password_generator"
	ActionExtractPDFImages  ActionType = "extract_pdf_images"
	ActionVideoToSticker    ActionType = "video_to_sticker"
	ActionImageToSticker    ActionType = "image_to_sticker"
	ActionVideoToGIF        ActionType = "video_to_gif"
	ActionImageToGIF        ActionType = "image_to_gif"
	ActionWatermarkImages   ActionType = "watermark_images"
)
