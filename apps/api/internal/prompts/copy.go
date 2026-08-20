package prompts

// 用户侧可见文案（危机转介、评估边界、教练开场兜底）。
// 改词只改本文件；前端经 GET /api/copy 读取，禁止再写一份。

// CrisisHelp 危机转介正文（教练回复、API help、前端横幅共用）。
const CrisisHelp = `如果你正在经历强烈的自我伤害念头，或担心自己可能伤害他人，请立刻寻求专业或紧急帮助：
· 中国大陆紧急情况请拨打 120；也可拨打当地心理援助热线（如各地 12320 / 心理援助热线）
· 联系身边可信的家人或朋友陪同就医
· 可预约线下持证心理咨询或精神科门诊
本产品是职场心理教练，不能替代持证心理咨询或精神科诊疗，也不会提供危机陪护。此刻不宜继续常规 AI 陪聊；请先照顾安全。`

// CrisisElevatedError 基线 crisisLevel=elevated 时拦截新建/继续疏导的短错误。
const CrisisElevatedError = `当前评估标出了需要被认真对待的痛苦信号，不宜继续常规 AI 陪聊。请先查看求助资源或预约人工支持。`

// AssessmentCrisisHeadline 基线 elevated 时报告标题（覆盖模型输出，保证口径一致）。
const AssessmentCrisisHeadline = `你标出了需要被认真对待的痛苦信号。建议优先寻求专业或紧急支持；产品内职场教练不适合处理危机时刻。`

// AssessmentCrisisSteps 基线 elevated 时的下一步。
var AssessmentCrisisSteps = []string{
	"查看危机求助资源并联系身边可信的人",
	"需要时预约私教通道，由人工确认时间",
}

// AssessmentBoundaryNote 评估报告非诊疗声明。
const AssessmentBoundaryNote = `本评估用于自我觉察与教练个性化，不是心理诊断，也不能替代持证咨询或精神科诊疗。`

// CoachMergeNote 注入综合档案末尾的合读说明（给模型，不直接给用户）。
const CoachMergeNote = `合读说明：画像看协作与压力习惯，心理评估（基线）看近两周压力与卡住点，快照看此刻状态。三者描述同一个人，须合在一起用；禁止诊断病名，禁止只引用其中一块。`

// CoachStartTurn 开场 user 指令：三源合读，禁止拆成互不相关的材料。
const CoachStartTurn = `请给出开场：合读三源档案（职场画像、心理评估/基线、快照；有则用、缺则跳过）后自然接住用户。禁止诊断病名；不要分别背诵三份材料，要合成同一个人来听。邀请用户用 2–3 句补充「发生了什么」。
若困扰分≥8 或评测危机偏高，语气放缓并轻提示专业支持 / 预约通道。`

// CoachReplyTurn 每轮回复 user 指令。
const CoachReplyTurn = `请按此顺序写 reply：
1) 若用户本轮有情绪/疲惫/压力：第一句必须共情（具体点出心理层面的感受，勿空泛「辛苦了」）；
2) 结合综合档案调整接话方式与侧重点（画像定节奏，评估定压力与目标，快照对齐此刻），勿把三源资料割裂；
3) 再澄清事实 vs 自我评判，或命名情绪与需求；
4) 再给 1–3 个选项，并收束到一个 24h 动作（可选附「下一句怎么说」）。
若需要过关训练，suggestGate 填 hr|interview|salary，否则空字符串。`

// CoachHeuristicSawProfile 无 LLM 时、已有任一源的开场。
const CoachHeuristicSawProfile = "\n\n我这边已经把你留下的职场画像、心理评估和此刻记录合在一起看了（有哪项用哪项；风格参考，不是诊断）。"

// CoachHeuristicHighDistress 困扰分偏高时的补充。
const CoachHeuristicHighDistress = `你现在的困扰分偏高。我们可以先把步子放慢；若持续很难受，也可预约私人心理辅导通道——我这边是职场教练，不能替代诊疗。`

// CoachHeuristicAskMore 有档案时的追问。
const CoachHeuristicAskMore = "\n\n想先补充一两句：最近具体发生了什么？还是直接从「今天想带走的那一点」聊起？"

// CoachHeuristicNoProfile 三源皆无时的开场续句。
const CoachHeuristicNoProfile = `可以先用两三句话告诉我：刚才实际发生了什么？以及你此刻最卡住的一个感受是什么？（我会帮你把事实和自我评判拆开。）`

// PublicCopy 给前端的用户侧文案。增删字段时同步 apps/web/src/copy.js。
func PublicCopy() map[string]any {
	return map[string]any{
		"crisisHelp":               CrisisHelp,
		"assessmentBoundaryNote":   AssessmentBoundaryNote,
		"assessmentCrisisHeadline": AssessmentCrisisHeadline,
		"assessmentCrisisSteps":    AssessmentCrisisSteps,
	}
}
