import { useState, useEffect } from 'react'
import { useParams, useLocation, Link } from 'react-router-dom'
import { getOrderByToken } from '../api'
import { Order } from '../types'

export default function OrderConfirmation() {
  const { orderNumber } = useParams<{ orderNumber: string }>()
  const location = useLocation()
  const { trackingToken, totalAmount } = location.state || {}
  const [order, setOrder] = useState<Order | null>(null)

  useEffect(() => {
    if (!trackingToken) return
    const poll = async () => {
      try {
        const data = await getOrderByToken(trackingToken)
        setOrder(data)
      } catch {
        /* ignore */
      }
    }
    poll()
    const id = setInterval(poll, 2000)
    return () => clearInterval(id)
  }, [trackingToken])

  return (
    <div className="order-confirmation">
      <div className="confirmation-icon">✓</div>
      <h1>Order Confirmed!</h1>
      <p className="order-number">Order Number: <strong>{orderNumber}</strong></p>
      {totalAmount && (
        <p className="order-total">Total: <strong>${totalAmount.toFixed(2)}</strong></p>
      )}

      {order?.source === 'mirrord' && (
        <div className="confirmation-mirrord-badge">
          <span className="source-badge source-mirrord" title={order.source_topic}>
            ⚡ Processed via Kafka (mirrord)
          </span>
          {order.source_topic && (
            <small className="source-topic-hint">Topic: {order.source_topic}</small>
          )}
        </div>
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
