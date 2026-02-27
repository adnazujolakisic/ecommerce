import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { Product, CartItem } from '../types'
import { getProduct, getInventory } from '../api'

interface ProductDetailProps {
  addToCart: (item: CartItem) => void
}

export default function ProductDetail({ addToCart }: ProductDetailProps) {
  const { id } = useParams<{ id: string }>()
  const [product, setProduct] = useState<Product | null>(null)
  const [stock, setStock] = useState<number>(0)
  const [quantity, setQuantity] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [added, setAdded] = useState(false)

  useEffect(() => {
    if (id) loadProduct(id)
  }, [id])

  const loadProduct = async (productId: string) => {
    setLoading(true)
    try {
      const [productData, inventoryData] = await Promise.all([
        getProduct(productId),
        getInventory(productId).catch(() => ({ stock_quantity: 0, reserved_quantity: 0 }))
      ])
      setProduct(productData)
      setStock(inventoryData.stock_quantity - inventoryData.reserved_quantity)
    } catch (err) {
      setError('Product not found')
    } finally {
      setLoading(false)
    }
  }

  const handleAddToCart = () => {
    if (!product) return
    addToCart({
      productId: product.id,
      productName: product.name,
      price: product.price,
      quantity,
      imageUrl: product.image_url,
    })
    setAdded(true)
    setTimeout(() => setAdded(false), 2000)
  }

  if (loading) return <div className="loading">Loading...</div>
  if (error || !product) return <div className="error">{error || 'Product not found'}</div>

  return (
    <div className="product-detail">
      <Link to="/" className="back-link">← Back to Products</Link>

      <div className="product-detail-content">
        <div className="product-detail-image">
          <img src={product.image_url} alt={product.name} />
        </div>

        <div className="product-detail-info">
          <span className="product-category-badge">{product.category}</span>
          <h1>{product.name}</h1>
          <p className="product-description">{product.description}</p>
          <p className="product-price">${product.price.toFixed(2)}</p>

          <div className="stock-info flex items-center gap-2">
            {stock > 0 ? (
              <span className="in-stock">In Stock ({stock} available)</span>
            ) : (
              <span className="out-of-stock">Out of Stock</span>
            )}
            <button
              type="button"
              className="text-sm underline text-muted-foreground hover:text-foreground"
              onClick={() => id && loadProduct(id)}
              disabled={loading}
            >
              Refresh
            </button>
          </div>

          {stock > 0 && (
            <div className="add-to-cart-section">
              <div className="quantity-selector">
                <button onClick={() => setQuantity(q => Math.max(1, q - 1))}>-</button>
                <span>{quantity}</span>
                <button onClick={() => setQuantity(q => Math.min(stock, q + 1))}>+</button>
              </div>
              <button
                className="add-to-cart-btn large"
                onClick={handleAddToCart}
              >
                {added ? 'Added!' : 'Add to Cart'}
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
