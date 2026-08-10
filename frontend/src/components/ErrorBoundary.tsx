import { Component, type ErrorInfo, type ReactNode } from 'react'
import { TriangleAlert } from 'lucide-react'
import { Button } from './ui/Button'

interface Props {
  children: ReactNode
}

interface State {
  error: Error | null
}

/**
 * Catches render-phase errors so a single broken view can never white-screen
 * the whole console. Users get a recovery path instead.
 */
export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    // Keep render failures observable in the console for operators.
    console.error('[ErrorBoundary]', error, info.componentStack)
  }

  private reset = () => this.setState({ error: null })

  render() {
    if (this.state.error) {
      return (
        <div className="flex h-dvh items-center justify-center bg-surface p-6">
          <div className="flex w-full max-w-md flex-col items-center gap-3 rounded-2xl border border-slate-200 bg-white p-10 text-center shadow-card">
            <span className="flex h-12 w-12 items-center justify-center rounded-xl bg-danger-50 text-danger-500">
              <TriangleAlert className="h-6 w-6" aria-hidden />
            </span>
            <h1 className="text-base font-semibold text-slate-900">Something went wrong</h1>
            <p className="text-sm text-slate-500">
              The console hit an unexpected error. Your data is safe — try reloading this view.
            </p>
            <div className="mt-2 flex gap-2">
              <Button variant="outline" onClick={() => window.location.reload()}>
                Reload console
              </Button>
              <Button onClick={this.reset}>Try again</Button>
            </div>
          </div>
        </div>
      )
    }
    return this.props.children
  }
}
