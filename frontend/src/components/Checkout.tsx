import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { CartItem, ShippingAddress } from '../types'
import { Button, buttonVariants } from '@metalbear/ui'
import { checkout } from '../api'

interface CheckoutProps {
  cart: CartItem[]
  clearCart: () => void
}

export default function Checkout({ cart, clearCart }: CheckoutProps) {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [formData, setFormData] = useState({
    email: 'demo@metalbear.com',
    name: 'Demo User',
    street: '123 Demo Street',
    city: 'San Francisco',
    state: 'CA',
    zipCode: '94102',
    country: 'USA',
  })

  const total = cart.reduce((sum, item) => sum + item.price * item.quantity, 0)

  if (cart.length === 0) {
    return (
      <div className="checkout-empty">
        <h2>Your cart is empty</h2>
        <Link to="/" className="continue-shopping-btn">Continue Shopping</Link>
      </div>
    )
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    const shippingAddress: ShippingAddress = {
      street: formData.street,
      city: formData.city,
      state: formData.state,
      zip_code: formData.zipCode,
      country: formData.country,
    }

    try {
      const response = await checkout(cart, formData.email, formData.name, shippingAddress)
      if (response.success && response.order_number) {
        clearCart()
        navigate(`/order-confirmation/${response.order_number}`, {
          state: {
            trackingToken: response.tracking_token,
            totalAmount: response.total_amount
          }
        })
      } else {
        setError(response.message || 'Checkout failed')
      }
    } catch (err: unknown) {
      const msg = err && typeof err === 'object' && 'response' in err
        ? (err as { response?: { data?: { message?: string }; status?: number } }).response?.data?.message
        : err instanceof Error ? err.message : 'Network or server error'
      setError(msg || 'Checkout failed. Please try again.')
      console.error('Checkout error:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    setFormData(prev => ({ ...prev, [e.target.name]: e.target.value }))
  }

  return (
    <div className="checkout-page">
      <h1>Checkout</h1>

      <div className="checkout-content">
        <form className="checkout-form" onSubmit={handleSubmit}>
          <h2>Contact Information</h2>
          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              type="email"
              id="email"
              name="email"
              value={formData.email}
              onChange={handleChange}
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="name">Full Name</label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              required
            />
          </div>

          <h2>Shipping Address</h2>
          <div className="form-group">
            <label htmlFor="street">Street Address</label>
            <input
              type="text"
              id="street"
              name="street"
              value={formData.street}
              onChange={handleChange}
              required
            />
          </div>
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="city">City</label>
              <input
                type="text"
                id="city"
                name="city"
                value={formData.city}
                onChange={handleChange}
                required
              />
            </div>
            <div className="form-group">
              <label htmlFor="state">State</label>
              <input
                type="text"
                id="state"
                name="state"
                value={formData.state}
                onChange={handleChange}
                required
              />
            </div>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="zipCode">ZIP Code</label>
              <input
                type="text"
                id="zipCode"
                name="zipCode"
                value={formData.zipCode}
                onChange={handleChange}
                required
              />
            </div>
            <div className="form-group">
              <label htmlFor="country">Country</label>
              <select
                id="country"
                name="country"
                value={formData.country}
                onChange={handleChange}
              >
                <option value="USA">United States</option>
                <option value="Canada">Canada</option>
                <option value="UK">United Kingdom</option>
              </select>
            </div>
          </div>

          {error && <div className="checkout-error">{error}</div>}

          <Button type="submit" className={buttonVariants({ variant: "brand-primary" })} disabled={loading} style={{ width: '100%' }}>
            {loading ? 'Processing...' : `Place Order - $${total.toFixed(2)}`}
          </Button>
        </form>

        <div className="order-summary">
          <h2>Order Summary</h2>
          {cart.map(item => (
            <div key={item.productId} className="summary-item">
              <span>{item.productName} x {item.quantity}</span>
              <span>${(item.price * item.quantity).toFixed(2)}</span>
            </div>
          ))}
          <div className="summary-total">
            <span>Total</span>
            <span>${total.toFixed(2)}</span>
          </div>
        </div>
      </div>
    </div>
  )
}
