import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { DataTable, type Column } from './DataTable'
import type { PageInfo } from '@/types/api'

interface Row {
  id: string
  name: string
}

const columns: Column<Row>[] = [
  { key: 'name', header: 'Name', sortField: 'name' },
  { key: 'id', header: 'ID' },
]

const pageInfo: PageInfo = { next_cursor: 'c2', has_more: true, limit: 2, total_count: 5 }

const rows: Row[] = [
  { id: '1', name: 'Ada' },
  { id: '2', name: 'Linus' },
]

describe('DataTable', () => {
  it('renders rows and headers', () => {
    render(<DataTable columns={columns} data={rows} rowKey={(r) => r.id} />)
    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('Ada')).toBeInTheDocument()
    expect(screen.getByText('Linus')).toBeInTheDocument()
  })

  it('renders a skeleton while loading', () => {
    const { container } = render(<DataTable columns={columns} data={[]} loading rowKey={(r) => r.id} />)
    expect(container.querySelector('.animate-shimmer')).not.toBeNull()
  })

  it('renders the friendly empty state', () => {
    render(
      <DataTable
        columns={columns}
        data={[]}
        rowKey={(r) => r.id}
        emptyTitle="No records found"
        emptyDescription="Try adjusting your search."
      />,
    )
    expect(screen.getByText('No records found')).toBeInTheDocument()
    expect(screen.getByText('Try adjusting your search.')).toBeInTheDocument()
  })

  it('renders the error state with a retry action', () => {
    const onRetry = vi.fn()
    render(<DataTable columns={columns} data={[]} error onRetry={onRetry} rowKey={(r) => r.id} />)
    fireEvent.click(screen.getByRole('button', { name: /try again/i }))
    expect(onRetry).toHaveBeenCalledOnce()
  })

  it('toggles sort directive on header click (desc → asc → desc)', () => {
    const onSort = vi.fn()
    render(<DataTable columns={columns} data={rows} sort="name:desc" onSort={onSort} rowKey={(r) => r.id} />)
    fireEvent.click(screen.getByRole('button', { name: /name/i }))
    expect(onSort).toHaveBeenCalledWith('name:asc')
  })

  it('switches to a new field with desc default', () => {
    const onSort = vi.fn()
    render(<DataTable columns={columns} data={rows} sort="id:asc" onSort={onSort} rowKey={(r) => r.id} />)
    fireEvent.click(screen.getByRole('button', { name: /name/i }))
    expect(onSort).toHaveBeenCalledWith('name:desc')
  })

  it('shows the pagination summary and honors next/previous', () => {
    const onNext = vi.fn()
    const onPrevious = vi.fn()
    render(
      <DataTable
        columns={columns}
        data={rows}
        rowKey={(r) => r.id}
        pageInfo={pageInfo}
        page={2}
        hasPrevious
        onNext={onNext}
        onPrevious={onPrevious}
      />,
    )
    expect(screen.getByText(/Showing 3–4 of 5/)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /next/i }))
    expect(onNext).toHaveBeenCalledOnce()
    fireEvent.click(screen.getByRole('button', { name: /previous/i }))
    expect(onPrevious).toHaveBeenCalledOnce()
  })

  it('disables Next when has_more is false', () => {
    const onNext = vi.fn()
    render(
      <DataTable
        columns={columns}
        data={rows}
        rowKey={(r) => r.id}
        pageInfo={{ ...pageInfo, has_more: false }}
        page={1}
        hasPrevious={false}
        onNext={onNext}
      />,
    )
    fireEvent.click(screen.getByRole('button', { name: /next/i }))
    expect(onNext).not.toHaveBeenCalled()
  })

  it('fires onRowClick when provided', () => {
    const onRowClick = vi.fn()
    render(<DataTable columns={columns} data={rows} rowKey={(r) => r.id} onRowClick={onRowClick} />)
    fireEvent.click(screen.getByText('Ada'))
    expect(onRowClick).toHaveBeenCalledWith(rows[0])
  })
})
