export interface Product {
  id: string
  name: string
  description: string
  price: number
  image_url: string
  category: string
  created_at: string
}

export interface CartItem {
  productId: string
  productName: string
  price: number
  quantity: number
  imageUrl: string
}

export interface ShippingAddress {
  street: string
  city: string
  state: string
  zip_code: string
  country: string
}

export interface Order {
  id: string
  order_number: string
  customer_email: string
  customer_name: string
  shipping_address: ShippingAddress
  total_amount: number
  status: string
  tracking_token: string
  items: OrderItem[]
  created_at: string
  updated_at: string
}

export interface OrderItem {
  id: string
  product_id: string
  product_name: string
  quantity: number
  price_at_time: number
}

export interface CheckoutResponse {
  success: boolean
  order_id?: string
  order_number?: string
  tracking_token?: string
  total_amount?: number
  message?: string
}
