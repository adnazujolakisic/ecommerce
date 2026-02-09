import axios from 'axios'
import { Product, Order, CheckoutResponse, CartItem, ShippingAddress } from './types'

const api = axios.create({ baseURL: '' })

export const getProducts = async (): Promise<Product[]> => {
  const { data } = await api.get('/api/products')
  return data
}

export const getProduct = async (id: string): Promise<Product> => {
  const { data } = await api.get(`/api/products/${id}`)
  return data
}

export const searchProducts = async (query: string): Promise<Product[]> => {
  const { data } = await api.get(`/api/products/search?q=${encodeURIComponent(query)}`)
  return data
}

export const getProductsByCategory = async (category: string): Promise<Product[]> => {
  const { data } = await api.get(`/api/products/category/${category}`)
  return data
}

export const getInventory = async (productId: string): Promise<{ stock_quantity: number; reserved_quantity: number }> => {
  const { data } = await api.get(`/api/inventory/${productId}`)
  return data
}

export const checkout = async (
  cart: CartItem[],
  customerEmail: string,
  customerName: string,
  shippingAddress: ShippingAddress
): Promise<CheckoutResponse> => {
  const { data } = await api.post('/api/checkout', {
    customer_email: customerEmail,
    customer_name: customerName,
    shipping_address: shippingAddress,
    items: cart.map(item => ({
      product_id: item.productId,
      product_name: item.productName,
      quantity: item.quantity,
      price: item.price,
    })),
  })
  return data
}

export const getOrder = async (id: string): Promise<Order> => {
  const { data } = await api.get(`/api/orders/${id}`)
  return data
}

export const getOrderByToken = async (token: string): Promise<Order> => {
  const { data } = await api.get(`/api/orders/track/${token}`)
  return data
}

export const getOrderStatus = async (id: string): Promise<{ status: string }> => {
  const { data } = await api.get(`/api/orders/${id}/status`)
  return data
}
