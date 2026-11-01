/** Closes an element when a click lands outside of it. */
export function clickOutside(node: HTMLElement, onOutside: () => void) {
  const handler = (event: MouseEvent) => {
    if (!node.contains(event.target as Node)) onOutside()
  }
  document.addEventListener('mousedown', handler)
  return {
    destroy() {
      document.removeEventListener('mousedown', handler)
    },
  }
}
