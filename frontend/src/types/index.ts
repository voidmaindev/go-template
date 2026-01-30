// API Response types
export interface ApiResponse<T> {
  success: boolean
  data: T
  request_id?: string
}

export interface ApiError {
  success: false
  error: {
    code: string
    message: string
    domain?: string
    details?: Array<{
      field: string
      code: string
      message: string
    }>
  }
  request_id?: string
}

export interface PaginatedData<T> {
  data: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
  has_more: boolean
}

// Auth types
export interface User {
  id: number
  email: string
  name: string
  email_verified_at: string | null
  is_self_registered: boolean
  has_password: boolean
  created_at: string
  updated_at: string
}

export interface TokenResponse {
  access_token: string
  refresh_token: string
  expires_at: number
  user: User
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
  role_codes?: string[]
}

export interface RefreshTokenRequest {
  refresh_token: string
}

export interface UpdateProfileRequest {
  name: string
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
}

// Self-registration & OAuth types
export interface SelfRegisterRequest {
  email: string
  password: string
  name: string
}

export interface SelfRegisterResponse {
  message: string
  user_id: number
}

export interface VerifyEmailRequest {
  token: string
}

export interface ResendVerificationRequest {
  email: string
}

export interface ForgotPasswordRequest {
  email: string
}

export interface ResetPasswordRequest {
  token: string
  new_password: string
}

export interface SetPasswordRequest {
  new_password: string
}

export type OAuthProvider = 'google' | 'facebook' | 'apple'

export interface OAuthTokenRequest {
  code: string
  state: string
  provider: OAuthProvider
}

export interface ExternalIdentity {
  id: number
  provider: OAuthProvider
  email: string
  created_at: string
}

export interface IdentitiesResponse {
  identities: ExternalIdentity[]
}

// Item types
export interface Item {
  id: number
  name: string
  description: string
  price: number
  created_at: string
  updated_at: string
}

export interface CreateItemRequest {
  name: string
  description?: string
  price: number
}

export interface UpdateItemRequest {
  name?: string
  description?: string
  price?: number
}

// Country types
export interface Country {
  id: number
  name: string
  code: string
  created_at: string
  updated_at: string
}

export interface CreateCountryRequest {
  name: string
  code: string
}

export interface UpdateCountryRequest {
  name?: string
  code?: string
}

// City types
export interface City {
  id: number
  name: string
  country_id: number
  country?: Country
  created_at: string
  updated_at: string
}

export interface CreateCityRequest {
  name: string
  country_id: number
}

export interface UpdateCityRequest {
  name?: string
  country_id?: number
}

// Document types
export interface DocumentItem {
  id: number
  document_id: number
  item_id: number
  item?: Item
  quantity: number
  price: number
  line_total: number
  created_at: string
  updated_at: string
}

export interface Document {
  id: number
  code: string
  city_id: number
  city?: City
  document_date: string
  total_amount: number
  items?: DocumentItem[]
  created_at: string
  updated_at: string
}

export interface CreateDocumentItemRequest {
  item_id: number
  quantity: number
  price: number
}

export interface CreateDocumentRequest {
  code: string
  city_id: number
  document_date: string
  items: CreateDocumentItemRequest[]
}

export interface UpdateDocumentRequest {
  code?: string
  city_id?: number
  document_date?: string
}

export interface UpdateDocumentItemRequest {
  quantity?: number
  price?: number
}

// RBAC types
export interface Permission {
  domain: string
  actions: string[]
}

export interface Role {
  id: number
  code: string
  name: string
  description: string
  is_system: boolean
  permissions?: Permission[]
  created_at: string
  updated_at: string
}

export interface CreateRoleRequest {
  code: string
  name: string
  description?: string
  permissions: Permission[]
}

export interface UpdateRolePermissionsRequest {
  permissions: Permission[]
}

export interface AssignRoleRequest {
  role_code: string
}

export interface UserRole {
  code: string
  name: string
}

export interface UserRolesResponse {
  user_id: number
  roles: UserRole[]
}

export interface DomainInfo {
  name: string
  is_protected: boolean
}

export interface DomainsResponse {
  domains: DomainInfo[]
}

export interface ActionsResponse {
  actions: string[]
}

// Query params
export interface QueryParams {
  page?: number
  page_size?: number
  sort?: string
  order?: 'asc' | 'desc'
  [key: string]: string | number | undefined
}
