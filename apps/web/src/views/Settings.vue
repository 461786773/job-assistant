<template>
  <section>
    <div class="hero">
      <h1>设置与说明</h1>
      <p>产品边界、隐私与危机资源。本产品是职场心理教练，不是临床诊疗。</p>
    </div>

    <section class="panel">
      <h2 class="section-title">产品定位</h2>
      <p>
        默认身份是<strong>职场心理教练</strong>：在求职、入职适应、晋升、沟通等节点，帮你稳住状态、理清选择、练好表达。人事
        / 面试 / 谈薪是可调用的过关训练工具。
      </p>
      <p class="muted">
        主要服务：<strong>应届大学生</strong>、<strong>刚入职场的小白</strong>、以及
        <strong>30 多岁压力较大的职场人</strong>。不做临床诊断，也不替代持证心理咨询。
      </p>
    </section>

    <section class="panel" style="margin-top: 16px">
      <h2 class="section-title">隐私</h2>
      <ul class="plain-list">
        <li>求职任务与教练会话按账号隔离，仅本人可见。</li>
        <li>心理健康相关记录默认仅本人可见；首发无组织端监控。</li>
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
