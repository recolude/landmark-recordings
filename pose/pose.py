# Original code found here:
# https://google.github.io/mediapipe/solutions/pose#python-solution-api

import cv2
import mediapipe as mp
import numpy as np
import os
import json

mp_drawing = mp.solutions.drawing_utils
mp_drawing_styles = mp.solutions.drawing_styles
mp_pose = mp.solutions.pose


BG_COLOR = (192, 192, 192)  # gray


def video_image_files(dir):
    num_files = len([name for name in os.listdir(
        dir) if os.path.isfile(os.path.join(dir, name))])
    file_paths = []

    for i in range(num_files):
        file_paths.append("frames/frame_{0}.png".format(str(i+1).zfill(4)))

    return file_paths


def write_frames(frames, out_path):

    data_out = []

    with mp_pose.Pose(
            static_image_mode=False,
            model_complexity=2,
            enable_segmentation=True,
            min_detection_confidence=0.5,
            smooth_segmentation=True) as pose:

        for idx, file in enumerate(frames):
            image = cv2.imread(file)
            results = pose.process(cv2.cvtColor(image, cv2.COLOR_BGR2RGB))

            if not results.pose_landmarks:
                continue

            entry = []
            i = 0
            for mark in results.pose_world_landmarks.landmark:
                entry.append({
                    "id": i,
                    "x": mark.x, 
                    "y": mark.y, 
                    "z": mark.z, 
                })
                i += 1

            data_out.append(entry)

    with open(out_path, 'w') as outfile:
        json.dump(data_out, outfile)


def process_frames(frames, out_path):

    with mp_pose.Pose(
            static_image_mode=False,
            model_complexity=2,
            enable_segmentation=True,
            min_detection_confidence=0.5,
            smooth_segmentation=True) as pose:

        for idx, file in enumerate(frames):
            image = cv2.imread(file)
            image_height, image_width, _ = image.shape
            # Convert the BGR image to RGB before processing.
            results = pose.process(cv2.cvtColor(image, cv2.COLOR_BGR2RGB))

            if not results.pose_landmarks:
                continue
            print(
                f'Nose coordinates: ('
                f'{results.pose_landmarks.landmark[mp_pose.PoseLandmark.NOSE].x * image_width}, '
                f'{results.pose_landmarks.landmark[mp_pose.PoseLandmark.NOSE].y * image_height})'
            )

            annotated_image = image.copy()
            # Draw segmentation on the image.
            # To improve segmentation around boundaries, consider applying a joint
            # bilateral filter to "results.segmentation_mask" with "image".
            condition = np.stack(
                (results.segmentation_mask,) * 3, axis=-1) > 0.1
            bg_image = np.zeros(image.shape, dtype=np.uint8)
            bg_image[:] = BG_COLOR
            annotated_image = np.where(condition, annotated_image, bg_image)
            # Draw pose landmarks on the image.
            mp_drawing.draw_landmarks(
                annotated_image,
                results.pose_landmarks,
                mp_pose.POSE_CONNECTIONS,
                landmark_drawing_spec=mp_drawing_styles.get_default_pose_landmarks_style())
            cv2.imwrite(os.path.join(
                out_path, f"frame_{str(idx + 1).zfill(4)}.png"), annotated_image)

            # landmark_count = 0
            # for x in results.pose_world_landmarks.landmark:
            #     print(x)

            # print(landmark_count)

            # print(mp_pose.POSE_CONNECTIONS)

            # Plot pose world landmarks.
            # mp_drawing.plot_landmarks(
            #     results.pose_world_landmarks, mp_pose.POSE_CONNECTIONS)


if __name__ == "__main__":
    write_frames(video_image_files("frames"), "out.json")
    # process_frames(video_image_files("frames"), "frames_out")
