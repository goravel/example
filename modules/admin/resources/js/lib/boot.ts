/** What the server hands the admin, from the data-boot attribute of #admin. */
interface Boot {
  appName: string
  basePath: string
}

function read(): Boot {
  const boot = document.getElementById('admin')?.dataset.boot
  if (!boot) throw new Error('[admin] #admin[data-boot] not found: the admin only runs on its own page')

  return JSON.parse(boot) as Boot
}

export const { appName, basePath } = read()
