import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Product, CartItem } from '../types'
import { getProducts, searchProducts, getProductsByCategory, getInventory } from '../api'
import { Card, CardContent, CardHeader, CardTitle, Button, Badge, SearchInput, Select, SelectContent, SelectItem, SelectTrigger, SelectValue, buttonVariants } from '@metalbear/ui'

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
            console.error(`Failed to fetch inventory for product ${product.id}:`, err)
            return {
              ...product,
              stock: 0 // Show "Out of stock" when fetch fails (e.g. no row in DB branch)
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
              stock: 0 // Show "Out of stock" when fetch fails (e.g. no row in DB branch)
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

  return (
    <div className="pt-20 pb-8 max-w-[1200px] mx-auto product-list-container">
      <div className="search-filters-bar flex gap-4 mb-8 flex-wrap justify-center">
        <div className="flex gap-2 items-center flex-nowrap">
          <SearchInput
            placeholder="Search products..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            onClear={() => setSearch('')}
            className="w-[600px] max-w-[640px]"
          />
          <Button className={buttonVariants({ variant: "brand" })} onClick={handleSearch}>Search</Button>
          <Select value={category} onValueChange={setCategory}>
          <SelectTrigger className="min-w-[200px]">
            <SelectValue placeholder="All Categories" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Categories</SelectItem>
            <SelectItem value="t-shirts">T-Shirts</SelectItem>
            <SelectItem value="hoodies">Hoodies</SelectItem>
            <SelectItem value="accessories">Accessories</SelectItem>
          </SelectContent>
        </Select>
        </div>
      </div>

      {error && (
        <div className="p-8 text-center text-destructive">{error}</div>
      )}
      {import.meta.env.VITE_PREVIEW_ENV === 'true' && (
        <p className="mb-4 text-muted-foreground">This is preview env</p>
      )}
      <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        {loading ? (
          <div className="col-span-full p-8 text-center">Loading products...</div>
        ) : (
        products.map(product => (
          <Card key={product.id} className="flex flex-col">
            <Link to={`/products/${product.id}`} className="text-inherit no-underline">
              <div className="aspect-[3/4] overflow-hidden flex items-center justify-center bg-muted">
                <img 
                  src={product.image_url} 
                  alt={product.name} 
                  className="w-full h-full object-cover" 
                />
              </div>
            </Link>
            <CardHeader>
              <CardTitle>
                <Link to={`/products/${product.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
                  {product.name}
                </Link>
              </CardTitle>
              <p className="text-sm text-muted-foreground my-2">{product.category}</p>
            </CardHeader>
            <CardContent className="flex flex-col gap-4 flex-1">
              <div className="flex justify-between items-center">
                <p className="text-xl font-bold m-0">
                  ${product.price.toFixed(2)}
                </p>
                {product.stock !== undefined ? (
                  product.stock > 0 ? (
                    <Badge style={{ backgroundColor: 'hsl(var(--brand-yellow))', color: '#222222' }}>
                      {product.stock} in stock
                    </Badge>
                  ) : (
                    <Badge>Out of stock</Badge>
                  )
                ) : (
                  <Badge>Loading...</Badge>
                )}
              </div>
              <Button
                className={`${buttonVariants({ variant: "brand-primary" })} w-full`}
                onClick={() => handleAddToCart(product)}
                disabled={product.stock === 0}
              >
                {product.stock === 0 ? 'Out of Stock' : 'Add to Cart'}
              </Button>
            </CardContent>
          </Card>
        ))
        )}
      </div>

      {!loading && products.length === 0 && (
        <p className="p-8 text-center">No products found</p>
      )}
    </div>
  )
}
