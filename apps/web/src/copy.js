import { ref } from 'vue'
import { api } from './api'

/** 与后端 prompts.PublicCopy 对齐；启动时 loadCopy，改词只改 Go 包。 */
export const crisisHelp = ref('')
export const assessmentBoundaryNote = ref('')
export const assessmentCrisisHeadline = ref('')

export async function loadCopy() {
  const data = await api.getCopy()
  crisisHelp.value = data.crisisHelp || ''
  assessmentBoundaryNote.value = data.assessmentBoundaryNote || ''
  assessmentCrisisHeadline.value = data.assessmentCrisisHeadline || ''
}
