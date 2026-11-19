// The admin's own entry, loaded after every module's dashboard script.
import './admin.css'

const widgets = document.getElementById('widgets')!
if (!widgets.children.length) widgets.innerHTML = '<p>No module added a widget.</p>'
