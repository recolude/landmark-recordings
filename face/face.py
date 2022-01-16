import cv2
import mediapipe as mp
import os
import json

mp_drawing = mp.solutions.drawing_utils
mp_drawing_styles = mp.solutions.drawing_styles
mp_face_mesh = mp.solutions.face_mesh


def video_image_files(dir):
    num_files = len([name for name in os.listdir(
        dir) if os.path.isfile(os.path.join(dir, name))])
    file_paths = []

    for i in range(num_files):
        file_paths.append("frames/frame_{0}.png".format(str(i+1).zfill(4)))

    return file_paths


def process_frames(frames, out_frame_path, out_path):
    data_out = []

    with mp_face_mesh.FaceMesh(
            static_image_mode=False,
            max_num_faces=2,
            refine_landmarks=True,
            min_detection_confidence=0.5) as face_mesh:

        drawing_spec = mp_drawing.DrawingSpec(thickness=1, circle_radius=1)
        for idx, file in enumerate(frames):
            image = cv2.imread(file)
            results = face_mesh.process(cv2.cvtColor(image, cv2.COLOR_BGR2RGB))

            # Print and draw face mesh landmarks on the image.
            if not results.multi_face_landmarks:
                continue

            entry = []

            annotated_image = image.copy()
            face_index = 0
            for face_landmarks in results.multi_face_landmarks:

                i = 0
                for mark in face_landmarks.landmark:
                    entry.append({
                        "id": i,
                        "face-id": face_index,
                        "x": mark.x,
                        "y": mark.y,
                        "z": mark.z,
                    })
                    i += 1

                data_out.append(entry)

                mp_drawing.draw_landmarks(
                    image=annotated_image,
                    landmark_list=face_landmarks,
                    connections=mp_face_mesh.FACEMESH_TESSELATION,
                    landmark_drawing_spec=None,
                    connection_drawing_spec=mp_drawing_styles
                    .get_default_face_mesh_tesselation_style())

                mp_drawing.draw_landmarks(
                    image=annotated_image,
                    landmark_list=face_landmarks,
                    connections=mp_face_mesh.FACEMESH_CONTOURS,
                    landmark_drawing_spec=None,
                    connection_drawing_spec=mp_drawing_styles
                    .get_default_face_mesh_contours_style())

                mp_drawing.draw_landmarks(
                    image=annotated_image,
                    landmark_list=face_landmarks,
                    connections=mp_face_mesh.FACEMESH_IRISES,
                    landmark_drawing_spec=None,
                    connection_drawing_spec=mp_drawing_styles
                    .get_default_face_mesh_iris_connections_style())

                face_index += 1

            out_img_path = os.path.join(
                out_frame_path, f"frame_{str(idx + 1).zfill(4)}.png")
            cv2.imwrite(out_img_path, annotated_image)

    with open(out_path, 'w') as outfile:
        json.dump(data_out, outfile)


if __name__ == "__main__":
    process_frames(video_image_files("frames"), "frames_out", "face.json")
