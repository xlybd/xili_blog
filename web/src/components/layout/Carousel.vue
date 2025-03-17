<template>
    <div class="carousel">
      <el-carousel trigger="click" height="700px">
        {

        }
        <el-carousel-item v-for="item in imgList" :key="item">
          <el-image fit="cover" :src=item alt=""></el-image>
        </el-carousel-item>
      </el-carousel>
    </div>
  </template>
  
  <script setup lang="ts">
  import {ref} from "vue";
  import {websiteCarousel} from "@/api/website";
  
  const imgList = ref<string[]>([
    '/image/carousel_1.jpg',
    '/image/carousel_2.jpg',
    '/image/carousel_3.jpg',
    '/image/carousel_4.jpg',
  ])
  
  const getWebsiteCarousel = async () => {
    const res = await websiteCarousel()
    if (res.code === 0 && res.data.length !== 0) {
      imgList.value = res.data
    }
  }
  
  getWebsiteCarousel()
  </script>
  
  <style scoped lang="scss">
  .carousel {
    width: 100%;
    position: relative;
  
    .el-carousel {
      .el-image {
        width: 100%;
        height: 100%;
      }
    }

    /* 修改轮播图左右切换按钮 */
  :deep(.el-carousel__arrow) {
    background-color: rgba(255, 255, 255, 0.3) !important; /* 半透明白色 */
    color: white !important; /* 箭头颜色 */
    border-radius: 50%; /* 按钮变圆 */
    width: 50px;
    height: 50px;
    transition: background-color 0.3s ease-in-out;
  }

  }
  </style>
  