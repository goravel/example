import { createSlot } from '../lib/slot'

/** The dashboard's widget grid. Modules fill it from the `fills` of their registration, in module order. */
export const DashboardWidgets = createSlot('dashboard-widgets')

/** The top right of the header. A page renders <HeaderActions.Fill> for as long as it is shown. */
export const HeaderActions = createSlot('header-actions')
