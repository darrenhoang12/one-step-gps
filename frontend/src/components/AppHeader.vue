<template>
  <header class="topbar">
    <div class="brand-block">
      <img
        class="brand-logo"
        src="@/assets/osg-logo.png"
        alt="OneStepGPS"
      />
    </div>
    <div class="topbar-right">
      <div class="avatar">OS</div>
    </div>
  </header>
</template>

<style scoped>
.topbar {
  position: relative;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 72px;
  padding: 0 28px;
  background: #12253c;
  color: #f6f8fb;
  box-shadow: 0 2px 15px #0d1c2b25;
}
.brand-block,
.topbar-right {
  display: flex;
  align-items: center;
}
.brand-logo {
  display: block;
  width: 190px;
  height: auto;
  padding: 7px 10px;
  border-radius: 10px;
  background: #fff;
}

.topbar-right {
  gap: 20px;
}
.demo-label {
  color: #cbd6e1;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.12em;
  display: flex;
  align-items: center;
  gap: 8px;
}
.demo-label i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #f1ba63;
  box-shadow: 0 0 0 4px #f1ba6324;
}
.avatar {
  display: grid;
  place-items: center;
  width: 35px;
  height: 35px;
  border: 1px solid #7b8c9e;
  border-radius: 50%;
  background: #42566d;
  color: #fff;
  font-size: 11px;
  font-weight: 800;
}
@media (max-width: 760px) {
  .topbar {
    height: 62px;
    padding: 0 16px;
  }
  .brand-logo {
    width: 158px;
    padding: 6px 9px;
  }
  .demo-label {
    display: none;
  }
}
</style>
