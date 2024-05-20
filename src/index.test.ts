import app from './index'

describe('Base', () => {
  test('GET /', async () => {
    const res = await app.request('http://localhost/')
    expect(res.status).toBe(200)
    expect(res.headers.get('x-powered-by')).toBe('Hono')
  })
})
