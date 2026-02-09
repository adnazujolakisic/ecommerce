import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Product, CartItem } from '../types'
import { getProducts, searchProducts, getProductsByCategory, getInventory } from '../api'

interface ProductListProps {
  addToCart: (item: CartItem) => void
}

interface ProductWithStock extends Product {
  stock?: number
}

export default function ProductList({ addToCart }: ProductListProps) {
  const [products, setProducts] = useState<ProductWithStock[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState<string>('all')

  useEffect(() => {
    loadProducts()
  }, [category])

  const loadProducts = async () => {
    setLoading(true)
    setError(null)
    try {
      let data: Product[]
      if (category === 'all') {
        data = await getProducts()
      } else {
        data = await getProductsByCategory(category)
      }
      
      // Fetch inventory for all products
      const productsWithStock = await Promise.all(
        data.map(async (product) => {
          try {
            const inventory = await getInventory(product.id)
            const availableStock = inventory.stock_quantity - inventory.reserved_quantity
            return {
              ...product,
              stock: Math.max(0, availableStock) // Ensure non-negative
            }
          } catch (err) {
            // If inventory fetch fails, log error but don't set stock to 0
            // This way we can distinguish between "no stock" and "failed to fetch"
            console.error(`Failed to fetch inventory for product ${product.id}:`, err)
            return {
              ...product,
              stock: undefined // Will show "Loading..." instead of "Out of stock"
            }
          }
        })
      )
      
      setProducts(productsWithStock)
    } catch (err) {
      setError('Failed to load products')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = async () => {
    if (!search.trim()) {
      loadProducts()
      return
    }
    setLoading(true)
    try {
      const data = await searchProducts(search)
      
      // Fetch inventory for search results
      const productsWithStock = await Promise.all(
        data.map(async (product) => {
          try {
            const inventory = await getInventory(product.id)
            const availableStock = inventory.stock_quantity - inventory.reserved_quantity
            return {
              ...product,
              stock: Math.max(0, availableStock) // Ensure non-negative
            }
          } catch (err) {
            console.error(`Failed to fetch inventory for product ${product.id}:`, err)
            return {
              ...product,
              stock: undefined // Will show "Loading..." instead of "Out of stock"
            }
          }
        })
      )
      
      setProducts(productsWithStock)
    } catch (err) {
      setError('Search failed')
    } finally {
      setLoading(false)
    }
  }

  const handleAddToCart = (product: Product) => {
    addToCart({
      productId: product.id,
      productName: product.name,
      price: product.price,
      quantity: 1,
      imageUrl: product.image_url,
    })
  }

  if (loading) return <div className="loading">Loading products...</div>
  if (error) return <div className="error">{error}</div>

  return (
    <div className="product-list-page">
      <div className="filters">
        <div className="search-bar">
          <input
            type="text"
            placeholder="Search products..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          />
          <button onClick={handleSearch}>Search</button>
        </div>
        <div className="category-filter">
          <select value={category} onChange={(e) => setCategory(e.target.value)}>
            <option value="all">All Categories</option>
            <option value="t-shirts">T-Shirts</option>
            <option value="hoodies">Hoodies</option>
            <option value="accessories">Accessories</option>
          </select>
        </div>
      </div>

      <div className="product-grid">
        {products.map(product => (
          <div key={product.id} className="product-card">
            <Link to={`/products/${product.id}`}>
              <div className="product-image">
                <img src={product.image_url} alt={product.name} />
              </div>
            </Link>
            <div className="product-info">
              <Link to={`/products/${product.id}`}>
                <h3>{product.name}</h3>
              </Link>
              <p className="product-category">{product.category}</p>
              <div className="product-price-row">
                <p className="product-price">${product.price.toFixed(2)}</p>
                <p className="product-stock">
                  {product.stock !== undefined ? (
                    product.stock > 0 ? (
                      <span className="in-stock-badge">{product.stock} in stock</span>
                    ) : (
                      <span className="out-of-stock-badge">Out of stock</span>
                    )
                  ) : (
                    <span className="stock-loading">Loading...</span>
                  )}
                </p>
              </div>
              <button
                className="add-to-cart-btn"
                onClick={() => handleAddToCart(product)}
                disabled={product.stock === 0}
              >
                {product.stock === 0 ? 'Out of Stock' : 'Add to Cart'}
              </button>
            </div>
          </div>
        ))}
      </div>

      {products.length === 0 && (
        <p className="no-products">No products found</p>
      )}
    </div>
  )
}
