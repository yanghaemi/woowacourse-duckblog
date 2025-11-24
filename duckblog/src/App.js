import React, { useState, useEffect } from 'react';
import { Music, Plus, Trash2, Edit2, X } from 'lucide-react';

export default function AlbumManager() {
  const [albums, setAlbums] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [formData, setFormData] = useState({
    title: '',
    artist: '',
    price: ''
  });

  const API_URL = 'http://localhost:8080';

  // 앨범 목록 불러오기
  const fetchAlbums = async () => {
    setIsLoading(true);
    try {
      const response = await fetch(`${API_URL}/albums`);
      const data = await response.json();
      setAlbums(data || []);
    } catch (error) {
      console.error('Failed to fetch albums:', error);
      alert('앨범을 불러오는데 실패했습니다.');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchAlbums();
  }, []);

  // 앨범 추가/수정
  const handleSubmit = async () => {
    if (!formData.title || !formData.artist || !formData.price) {
      alert('모든 필드를 입력해주세요.');
      return;
    }
    
    const albumData = {
      title: formData.title,
      artist: formData.artist,
      price: parseFloat(formData.price)
    };

    try {
      if (editingId) {
        // 수정
        await fetch(`${API_URL}/albums/${editingId}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(albumData)
        });
      } else {
        // 추가
        await fetch(`${API_URL}/albums`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(albumData)
        });
      }
      
      fetchAlbums();
      resetForm();
    } catch (error) {
      console.error('Failed to save album:', error);
      alert('저장에 실패했습니다.');
    }
  };

  // 앨범 삭제
  const handleDelete = async (id) => {
    if (!window.confirm('정말 삭제하시겠습니까?')) return;
    
    try {
      await fetch(`${API_URL}/albums/${id}`, {
        method: 'DELETE'
      });
      fetchAlbums();
    } catch (error) {
      console.error('Failed to delete album:', error);
      alert('삭제에 실패했습니다.');
    }
  };

  // 수정 모드
  const handleEdit = (album) => {
    setEditingId(album.id);
    setFormData({
      title: album.title,
      artist: album.artist,
      price: album.price.toString()
    });
    setShowForm(true);
  };

  // 폼 초기화
  const resetForm = () => {
    setFormData({ title: '', artist: '', price: '' });
    setEditingId(null);
    setShowForm(false);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 to-blue-50 p-8">
      <div className="max-w-4xl mx-auto">
        {/* 헤더 */}
        <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Music className="w-8 h-8 text-purple-600" />
              <h1 className="text-3xl font-bold text-gray-800">Album Manager</h1>
            </div>
            <button
              onClick={() => setShowForm(!showForm)}
              className="flex items-center gap-2 bg-purple-600 text-white px-4 py-2 rounded-lg hover:bg-purple-700 transition"
            >
              {showForm ? <X className="w-5 h-5" /> : <Plus className="w-5 h-5" />}
              {showForm ? '취소' : '앨범 추가'}
            </button>
          </div>
        </div>

        {/* 폼 */}
        {showForm && (
          <div className="bg-white rounded-lg shadow-lg p-6 mb-6">
            <h2 className="text-xl font-semibold mb-4 text-gray-800">
              {editingId ? '앨범 수정' : '새 앨범 추가'}
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  제목
                </label>
                <input
                  type="text"
                  value={formData.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  아티스트
                </label>
                <input
                  type="text"
                  value={formData.artist}
                  onChange={(e) => setFormData({ ...formData, artist: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  가격
                </label>
                <input
                  type="number"
                  step="0.01"
                  value={formData.price}
                  onChange={(e) => setFormData({ ...formData, price: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                />
              </div>
              <div className="flex gap-2">
                <button
                  onClick={handleSubmit}
                  className="flex-1 bg-purple-600 text-white py-2 rounded-lg hover:bg-purple-700 transition font-medium"
                >
                  {editingId ? '수정' : '추가'}
                </button>
                <button
                  onClick={resetForm}
                  className="px-6 bg-gray-200 text-gray-700 py-2 rounded-lg hover:bg-gray-300 transition"
                >
                  취소
                </button>
              </div>
            </div>
          </div>
        )}

        {/* 앨범 리스트 */}
        <div className="bg-white rounded-lg shadow-lg p-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-800">
            앨범 목록 ({albums.length})
          </h2>
          
          {isLoading ? (
            <div className="text-center py-8 text-gray-500">로딩 중...</div>
          ) : albums.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              앨범이 없습니다. 첫 앨범을 추가해보세요!
            </div>
          ) : (
            <div className="space-y-3">
              {albums.map((album) => (
                <div
                  key={album.id}
                  className="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition"
                >
                  <div className="flex-1">
                    <h3 className="font-semibold text-gray-800">{album.title}</h3>
                    <p className="text-sm text-gray-600">{album.artist}</p>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className="text-lg font-bold text-purple-600">
                      ${album.price.toFixed(2)}
                    </span>
                    <div className="flex gap-2">
                      <button
                        onClick={() => handleEdit(album)}
                        className="p-2 text-blue-600 hover:bg-blue-50 rounded-lg transition"
                      >
                        <Edit2 className="w-5 h-5" />
                      </button>
                      <button
                        onClick={() => handleDelete(album.id)}
                        className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition"
                      >
                        <Trash2 className="w-5 h-5" />
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}