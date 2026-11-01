<script lang="ts">
  import { chart } from '../lib/charts'

  export let config: any
  export let height = 220
</script>

<div class="chart" style="height:{height}px">
  <canvas use:chart={config}></canvas>
</div>

<style>
  .chart {
    position: relative;
    width: 100%;
  }
</style>
