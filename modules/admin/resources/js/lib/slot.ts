import { defineComponent, h, inject, provide, shallowRef, Teleport, type Component, type InjectionKey, type ShallowRef } from 'vue'

export interface Slot {
  /** Wraps both the outlet and every fill. */
  Provider: Component
  /** The spot the layout owns; fills are teleported into it in mount order. */
  Outlet: Component
  /** Places its children into the outlet while both are mounted; renders nothing otherwise. */
  Fill: Component
}

export function createSlot(name: string): Slot {
  const key: InjectionKey<ShallowRef<HTMLElement | null>> = Symbol(name)

  function useTarget(): ShallowRef<HTMLElement | null> {
    const target = inject(key, null)
    if (!target) throw new Error(`[slot] "${name}" used outside its provider`)

    return target
  }

  return {
    Provider: defineComponent({
      name: `${name}-provider`,
      setup(_, { slots }) {
        provide(key, shallowRef<HTMLElement | null>(null))

        return () => slots.default?.()
      },
    }),

    Outlet: defineComponent({
      name: `${name}-outlet`,
      setup() {
        const target = useTarget()
        // One function for the component's life, so Vue calls it on mount and unmount only.
        const setTarget = (element: unknown) => {
          target.value = element as HTMLElement | null
        }

        return () => h('div', { ref: setTarget, 'data-slot': name })
      },
    }),

    Fill: defineComponent({
      name: `${name}-fill`,
      setup(_, { slots }) {
        const target = useTarget()

        return () => (target.value ? h(Teleport, { to: target.value }, slots.default?.() ?? []) : null)
      },
    }),
  }
}
