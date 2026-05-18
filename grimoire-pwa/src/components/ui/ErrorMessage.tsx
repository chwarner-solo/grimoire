interface Props {
  message: string
}

export function ErrorMessage({ message }: Props) {
  return (
    <div className="rounded-md bg-red-950 border border-red-800 p-4 text-sm text-red-300">
      {message}
    </div>
  )
}
