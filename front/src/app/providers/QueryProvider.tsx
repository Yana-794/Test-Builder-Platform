import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,          // 30 сек — данные считаются свежими
      gcTime: 5 * 60_000,         // 5 мин — хранится в кеше после unmount
      retry: 1,                   // 1 повтор при ошибке
      refetchOnWindowFocus: false, // не дёргать запросы при возврате в таб
    },
    mutations: {
      retry: 0,
    },
  },
})

type Props = { children: ReactNode }

export function QueryProvider({ children }: Props) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}