import { useState, useEffect, useRef } from 'react'
import { useParams, Link } from 'react-router-dom'
import { Order } from '../types'
import { getOrderByToken } from '../api'

const statusSteps = ['pending', 'processing', 'confirmed', 'shipped', 'delivered']

export default function OrderTracking() {
  const { token } = useParams<{ token: string }>()
  const [order, setOrder] = useState<Order | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const isInitialLoad = useRef(true)
  const pollIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    if (!token) return
    
    // Initial load
    const loadInitial = async () => {
      setLoading(true)
      try {
        const data = await getOrderByToken(token)
        setOrder(data)
        setError(null)
        isInitialLoad.current = false
      } catch (err) {
        setError('Order not found')
      } finally {
        setLoading(false)
      }
    }
    
    loadInitial()
    
    // Poll for status updates every 2 seconds
    // Stop polling if order is in a final state
    const finalStates = ['shipped', 'delivered', 'cancelled']
    
    pollIntervalRef.current = setInterval(async () => {
      try {
        const data = await getOrderByToken(token)
        console.log('Polling order status:', data.status, data.order_number)
        
        // Always update the order to trigger re-render
        setOrder(data)
        
        // Stop polling if order reaches a final state
        if (finalStates.includes(data.status)) {
          console.log('Order reached final state, stopping polling')
          if (pollIntervalRef.current) {
            clearInterval(pollIntervalRef.current)
            pollIntervalRef.current = null
          }
        }
      } catch (err) {
        console.error('Failed to poll order status:', err)
        // Continue polling even on error
      }
    }, 2000)
    
    return () => {
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current)
        pollIntervalRef.current = null
      }
    }
  }, [token])

  if (loading) return <div className="loading">Loading order...</div>
  if (error || !order) return <div className="error">{error || 'Order not found'}</div>

  const currentStepIndex = statusSteps.indexOf(order.status)

  return (
    <div className="order-tracking">
      <h1>Order Tracking</h1>
      <p className="order-number">Order Number: <strong>{order.order_number}</strong></p>

      <div className="tracking-progress">
        {statusSteps.map((step, index) => (
          <div
            key={step}
            className={`tracking-step ${index <= currentStepIndex ? 'completed' : ''} ${index === currentStepIndex ? 'current' : ''}`}
          >
            <div className="step-indicator">{index <= currentStepIndex ? '✓' : index + 1}</div>
            <span className="step-label">{step.charAt(0).toUpperCase() + step.slice(1)}</span>
          </div>
        ))}
      </div>

      <div className="order-details">
        <h2>Order Details</h2>
        <div className="detail-row">
          <span>Status:</span>
          <span className={`status-badge status-${order.status}`}>{order.status}</span>
        </div>
        <div className="detail-row">
          <span>Total:</span>
          <span>${order.total_amount.toFixed(2)}</span>
        </div>
        <div className="detail-row">
          <span>Ordered:</span>
          <span>{new Date(order.created_at).toLocaleDateString()}</span>
        </div>

        <h3>Items</h3>
        <div className="order-items">
          {order.items?.map(item => (
            <div key={item.id} className="order-item">
              <span>{item.product_name} x {item.quantity}</span>
              <span>${(item.price_at_time * item.quantity).toFixed(2)}</span>
            </div>
          ))}
        </div>

        <h3>Shipping Address</h3>
        <div className="shipping-address">
          <p>{order.customer_name}</p>
          <p>{order.shipping_address.street}</p>
          <p>{order.shipping_address.city}, {order.shipping_address.state} {order.shipping_address.zip_code}</p>
          <p>{order.shipping_address.country}</p>
        </div>
      </div>

      <Link to="/" className="continue-shopping-link">
        Continue Shopping
      </Link>
    </div>
  )
}
