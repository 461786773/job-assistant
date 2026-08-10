<template>
  <section>
    <div class="hero">
      <h1>设置与说明</h1>
      <p>产品边界、隐私与危机资源。本产品是职场心理教练，不是临床诊疗。</p>
    </div>

    <section class="panel">
      <h2 class="section-title">产品定位</h2>
      <p>默认身份是<strong>职场心理教练</strong>：在求职、晋升、沟通节点稳住状态、理清选择、练好表达。人事 / 业务 / 谈薪是可调用的过关训练工具。</p>
      <p class="muted">服务对象优先：3–10 年 ToB / 安全 / 合规迁移或晋升中的职场人。不做应届生、纯小白路径；不做临床诊断。</p>
    </section>

    <section class="panel" style="margin-top: 16px">
      <h2 class="section-title">隐私</h2>
      <ul class="plain-list">
        <li>求职任务与教练会话按用户名隔离，仅本人可见。</li>
        <li>心理健康打卡默认仅本人可见；首发无组织端监控。</li>
        <li>数据用于教练个性化与趋势对照，不作录用/晋升评判依据。</li>
      </ul>
    </section>

    <section class="panel" style="margin-top: 16px">
      <h2 class="section-title">危机时请这样</h2>
      <p class="crisis-banner" style="margin: 0">{{ CRISIS_HELP }}</p>
    </section>
  </section>
</template>

<script setup>
import { CRISIS_HELP } from '../api'
</script>
