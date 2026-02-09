import { useParams, useLocation, Link } from 'react-router-dom'

export default function OrderConfirmation() {
  const { orderNumber } = useParams<{ orderNumber: string }>()
  const location = useLocation()
  const { trackingToken, totalAmount } = location.state || {}

  return (
    <div className="order-confirmation">
      <div className="confirmation-icon">✓</div>
      <h1>Order Confirmed!</h1>
      <p className="order-number">Order Number: <strong>{orderNumber}</strong></p>
      {totalAmount && (
        <p className="order-total">Total: <strong>${totalAmount.toFixed(2)}</strong></p>
      )}

      <div className="confirmation-details">
        <p>Thank you for your purchase! You will receive an email confirmation shortly.</p>
        <p>Your order is being processed and will be shipped soon.</p>
      </div>

      {trackingToken && (
        <Link to={`/track/${trackingToken}`} className="track-order-btn">
          Track Your Order
        </Link>
      )}

      <Link to="/" className="continue-shopping-link">
        Continue Shopping
      </Link>
    </div>
  )
}
